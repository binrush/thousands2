#!/bin/bash
set -euo pipefail

# Backup script for Thousands2 database
# This script is executed before the service starts
# Usage: backup.sh <backup_dir> <db_path>

if [ $# -ne 2 ]; then
    echo "Usage: $0 <backup_dir> <db_path>" >&2
    exit 1
fi

BACKUP_DIR="$1"
DB_PATH="$2"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/thousands_backup_$TIMESTAMP.sql.gz"

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

# Create backup - dump SQLite database and compress with gzip
sqlite3 "$DB_PATH" .dump | gzip > "$BACKUP_FILE"

echo "Backup created successfully: $BACKUP_FILE"

# Delete backups older than 7 days
find "$BACKUP_DIR" -name "thousands_backup_*.sql.gz" -type f -mtime +7 -delete

echo "Cleanup completed: old backups (7+ days) removed"

