package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockS3Service simulates S3 behavior for testing
type MockS3Service struct {
	objects map[string][]byte
}

func NewMockS3Service() *MockS3Service {
	return &MockS3Service{
		objects: make(map[string][]byte),
	}
}

func (m *MockS3Service) HandlePutObject(w http.ResponseWriter, r *http.Request) {
	// Extract bucket and key from URL path
	// Expected format: /{bucket}/{key}
	path := r.URL.Path[1:] // Remove leading slash
	if path == "" {
		http.Error(w, "Missing bucket and key", http.StatusBadRequest)
		return
	}

	// For simplicity, assume bucket is always the first part and key is the rest
	key := path[len(S3_BUCKET)+1:] // Remove bucket name + slash
	if key[0] == '/' {
		key = key[1:] // Remove leading slash from key
	}

	// Read the body
	body := make([]byte, r.ContentLength)
	_, err := r.Body.Read(body)
	if err != nil && err.Error() != "EOF" {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	// Store the object
	m.objects[key] = body

	// Return success response
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func (m *MockS3Service) HandleGetObject(w http.ResponseWriter, r *http.Request) {
	// Extract key from URL path
	path := r.URL.Path[1:]         // Remove leading slash
	key := path[len(S3_BUCKET)+1:] // Remove bucket name + slash
	if key[0] == '/' {
		key = key[1:] // Remove leading slash from key
	}

	// Check if object exists
	object, exists := m.objects[key]
	if !exists {
		http.Error(w, "Object not found", http.StatusNotFound)
		return
	}

	// Return the object
	w.Header().Set("Content-Type", "image/jpeg")
	w.Write(object)
}

func (m *MockS3Service) HandleListObjects(w http.ResponseWriter, r *http.Request) {
	// Return a simple XML response for list objects
	response := `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult xmlns="http://s3.amazonaws.com/doc/2006-03-01/">
	<Name>` + S3_BUCKET + `</Name>
	<Prefix></Prefix>
	<Marker></Marker>
	<MaxKeys>1000</MaxKeys>
	<IsTruncated>false</IsTruncated>
</ListBucketResult>`

	w.Header().Set("Content-Type", "application/xml")
	w.Write([]byte(response))
}

func (m *MockS3Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "PUT":
		m.HandlePutObject(w, r)
	case "GET":
		if r.URL.Query().Get("list-type") == "2" {
			m.HandleListObjects(w, r)
		} else {
			m.HandleGetObject(w, r)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetObjectStored returns the stored object data for verification
func (m *MockS3Service) GetObjectStored(key string) ([]byte, bool) {
	data, exists := m.objects[key]
	return data, exists
}

func TestS3ImageManagerUpload(t *testing.T) {
	// Create mock S3 server
	mockS3 := NewMockS3Service()
	server := httptest.NewServer(mockS3)
	defer server.Close()

	ctx := context.Background()

	tests := []struct {
		name        string
		imageData   []byte
		key         string
		expectError bool
	}{
		{
			name:        "successful upload",
			imageData:   []byte("fake image data"),
			key:         "test/image.jpg",
			expectError: false,
		},
		{
			name:        "empty image data",
			imageData:   []byte(""),
			key:         "test/empty.jpg",
			expectError: false, // S3 allows empty objects
		},
		{
			name:        "key with special characters",
			imageData:   []byte("fake image data"),
			key:         "test/images/2024/01/image with spaces & symbols!.jpg",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create S3ImageManager with mock endpoint
			manager, err := NewS3ImageManager("test-access-key", "test-secret-key", server.URL, ctx)
			assert.NoError(t, err)
			assert.NotNil(t, manager)

			// Test upload
			err = manager.Upload(ctx, tt.imageData, tt.key)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Verify the object was stored
				storedData, exists := mockS3.GetObjectStored(tt.key)
				assert.True(t, exists, "Object should be stored")
				assert.Equal(t, tt.imageData, storedData, "Stored data should match input")
			}
		})
	}
}

func TestS3ImageManagerUploadWithContextCancellation(t *testing.T) {
	// Create mock S3 server
	mockS3 := NewMockS3Service()
	server := httptest.NewServer(mockS3)
	defer server.Close()

	ctx := context.Background()

	// Create S3ImageManager with mock endpoint
	manager, err := NewS3ImageManager("test-access-key", "test-secret-key", server.URL, ctx)
	assert.NoError(t, err)

	// Create a cancelled context
	cancelledCtx, cancel := context.WithCancel(ctx)
	cancel()

	// Upload should fail due to cancelled context
	err = manager.Upload(cancelledCtx, []byte("test data"), "test.jpg")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
}
