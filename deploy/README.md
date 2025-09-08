# Thousands2 Deployment

This directory contains Ansible playbooks for deploying the Thousands2 application to a Debian server.

## Prerequisites

1. Ansible installed on your local machine
2. SSH access to the target server
3. Domain name pointing to the server's IP address
4. OAuth credentials for VK

## Configuration

1. Edit `inventory/hosts`:
   - Replace `your.server.ip` with your server's IP address
   - Replace `your.domain.com` with your domain name
   - Set your VK client ID

2. Configure SSL (optional):
   - Edit `group_vars/all.yml` to set `enable_ssl: false` if you want to disable SSL
   - By default, SSL is enabled and managed by Certbot
   - When SSL is disabled, the application will run on HTTP only

3. Set up secrets:
   ```bash
   # Create host-specific vault files
   ansible-vault create host_vars/alpha.yml
   ansible-vault create host_vars/beta.yml
   
   # Or edit existing vault files
   ansible-vault edit host_vars/alpha.yml
   ansible-vault edit host_vars/beta.yml
   ```

3. Add the following to each host's vault file:
   ```yaml
   vk_client_secret: "your_host_specific_vk_client_secret"
   ```

## Deployment

1. Build the application:
   ```bash
   ./build.sh
   ```

2. Deploy to the server:
   ```bash
   # Using vault password file
   ansible-playbook -i inventory/hosts deploy.yml --vault-password-file ~/.vault_pass.txt
   
   # Or prompt for vault password
   ansible-playbook -i inventory/hosts deploy.yml --ask-vault-pass
   
   # Deploy to specific host
   ansible-playbook -i inventory/hosts deploy.yml -l alpha --ask-vault-pass
   ```

## Directory Structure

- `/opt/thousands2/` - Application directory
  - `thousands2` - Application binary
  - `data/` - Data directory
  - `db/` - Database directory

## Security

- The application runs as a non-root user `thousands2`
- SSL certificates are automatically managed by Certbot (when enabled)
- OAuth credentials are passed via environment variables
- Automatic security updates are enabled
- Sensitive variables are encrypted using Ansible Vault
- Each host can have its own OAuth secret
- SSL can be disabled by setting `enable_ssl: false` in `group_vars/all.yml`

## Maintenance

- SSL certificates are automatically renewed (when SSL is enabled)
- The application automatically restarts on failure
- Logs can be viewed with `journalctl -u thousands2` 