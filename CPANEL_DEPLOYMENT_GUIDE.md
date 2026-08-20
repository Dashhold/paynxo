# PayNXO cPanel Deployment Guide

Complete guide to deploy PayNXO (Backend, Frontend, and Mobile App) on cPanel hosting.

---

## 🎯 Important Notes

**cPanel Limitations:**
- Most shared cPanel hosting doesn't support Go applications natively
- You'll need VPS or dedicated hosting with cPanel + SSH access
- Node.js apps require cPanel Node.js selector or SSH access

**Requirements:**
- cPanel with SSH access (root or sudo)
- Node.js support (cPanel Node.js selector)
- MySQL/PostgreSQL database
- SSL certificate (AutoSSL or Let's Encrypt)
- Domain: paynxo.com

---

## 📋 Prerequisites

### Server Requirements
- **cPanel Version**: 11.102 or higher
- **SSH Access**: Required for backend deployment
- **Node.js**: Version 18+ (via cPanel Node.js selector)
- **Database**: MySQL 5.7+ or PostgreSQL 12+
- **RAM**: Minimum 2GB
- **Storage**: Minimum 10GB

---

## 🚀 Deployment Steps

### PART 1: cPanel Setup

#### Step 1.1: Login to cPanel

1. Go to: `https://yourdomain.com:2083`
2. Login with your cPanel credentials
3. Navigate to the cPanel dashboard

---

### PART 2: Database Setup

#### Step 2.1: Create MySQL Database

1. **In cPanel, find "MySQL Databases"**
   - Click on "MySQL Database Wizard" or "MySQL Databases"

2. **Create Database:**
   - Database Name: `paymentgateway`
   - Click "Create Database"
   - Note the full database name (usually: `username_paymentgateway`)

3. **Create Database User:**
   - Username: `paynxo`
   - Password: Generate strong password
   - Click "Create User"
   - Note the full username (usually: `username_paynxo`)

4. **Grant Privileges:**
   - Select the database and user
   - Check "ALL PRIVILEGES"
   - Click "Make Changes"

5. **Save Credentials:**
   ```
   Database Name: username_paymentgateway
   Database User: username_paynxo
   Database Pass: [your generated password]
   Database Host: localhost
   ```

---

### PART 3: Frontend Deployment

#### Step 3.1: Build Frontend Locally

On your local machine (Windows):

```bash
cd C:\paynxo\frontend

# Install dependencies
npm install

# Build for production
npm run build

# This creates a 'dist' folder with all files
```

#### Step 3.2: Upload Frontend via cPanel File Manager

1. **In cPanel, open "File Manager"**
   
2. **Navigate to public_html:**
   - If paynxo.com is your main domain: `/public_html/`
   - If paynxo.com is addon domain: `/public_html/paynxo.com/`

3. **Upload Files:**
   - Click "Upload"
   - Upload ALL files from `C:\paynxo\frontend\dist\`
   - OR compress dist folder to zip and upload, then extract

4. **Verify Structure:**
   ```
   public_html/
   ├── index.html
   ├── assets/
   │   ├── index-[hash].css
   │   └── index-[hash].js
   └── favicon.ico (if exists)
   ```

#### Step 3.3: Configure .htaccess for SPA Routing

Create `.htaccess` in `public_html/`:

```apache
<IfModule mod_rewrite.c>
  RewriteEngine On
  RewriteBase /
  
  # Don't rewrite files or directories
  RewriteCond %{REQUEST_FILENAME} !-f
  RewriteCond %{REQUEST_FILENAME} !-d
  
  # Rewrite everything else to index.html to allow SPA routing
  RewriteRule ^ index.html [L]
</IfModule>

# Enable GZIP compression
<IfModule mod_deflate.c>
  AddOutputFilterByType DEFLATE text/html text/plain text/xml text/css text/javascript application/javascript application/json
</IfModule>

# Cache static assets
<IfModule mod_expires.c>
  ExpiresActive On
  ExpiresByType image/jpg "access plus 1 year"
  ExpiresByType image/jpeg "access plus 1 year"
  ExpiresByType image/gif "access plus 1 year"
  ExpiresByType image/png "access plus 1 year"
  ExpiresByType image/svg+xml "access plus 1 year"
  ExpiresByType text/css "access plus 1 year"
  ExpiresByType application/javascript "access plus 1 year"
  ExpiresByType application/x-javascript "access plus 1 year"
  ExpiresByType text/javascript "access plus 1 year"
</IfModule>

# Security Headers
<IfModule mod_headers.c>
  Header set X-Content-Type-Options "nosniff"
  Header set X-Frame-Options "SAMEORIGIN"
  Header set X-XSS-Protection "1; mode=block"
</IfModule>
```

---

### PART 4: Backend Deployment (Go Application)

**Note:** Go applications require SSH access and are not natively supported in shared cPanel hosting. You need VPS/Dedicated hosting.

#### Step 4.1: Connect via SSH

```bash
# From your local machine
ssh username@yourdomain.com -p 22

# Or use cPanel Terminal (if available)
```

#### Step 4.2: Install Go (if not installed)

```bash
# Check if Go is installed
go version

# If not installed:
cd ~
wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
tar -xzf go1.21.6.linux-amd64.tar.gz
echo 'export PATH=$HOME/go/bin:$PATH' >> ~/.bashrc
source ~/.bashrc
go version
```

#### Step 4.3: Clone Repository

```bash
# Navigate to home directory
cd ~

# Create app directory
mkdir -p apps/paynxo
cd apps/paynxo

# Clone repository
git clone https://github.com/Dashhold/paynxo.git .

# Or upload via cPanel File Manager and extract
```

#### Step 4.4: Configure Backend

```bash
cd ~/apps/paynxo/backend

# Create .env file
cat > .env << 'EOF'
# Server Configuration
PORT=8080
GIN_MODE=release

# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_NAME=username_paymentgateway
DB_USER=username_paynxo
DB_PASSWORD=your_database_password
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=your_jwt_secret_change_this_to_random_string
JWT_EXPIRY=24h

# CORS Configuration
ALLOWED_ORIGINS=https://paynxo.com,https://www.paynxo.com

# Application
APP_ENV=production
EOF

# Secure the file
chmod 600 .env
```

**Important:** Update these values:
- `DB_NAME`: Your actual database name from Step 2.1
- `DB_USER`: Your actual database user from Step 2.1
- `DB_PASSWORD`: Your database password from Step 2.1
- `JWT_SECRET`: Generate random string (use: `openssl rand -base64 64`)

#### Step 4.5: Build Backend

```bash
cd ~/apps/paynxo/backend

# Build the application
go build -o paynxo-api ./cmd/server

# Test run (should start on port 8080)
./paynxo-api

# Press Ctrl+C to stop
```

#### Step 4.6: Setup Process Manager (PM2 or systemd)

**Option A: Using PM2 (Recommended for cPanel)**

```bash
# Install PM2 globally
npm install -g pm2

# Create PM2 ecosystem file
cat > ~/apps/paynxo/backend/ecosystem.config.js << 'EOF'
module.exports = {
  apps: [{
    name: 'paynxo-api',
    script: './paynxo-api',
    cwd: '/home/username/apps/paynxo/backend',
    instances: 1,
    autorestart: true,
    watch: false,
    max_memory_restart: '1G',
    env: {
      NODE_ENV: 'production'
    }
  }]
};
EOF

# Start with PM2
cd ~/apps/paynxo/backend
pm2 start ecosystem.config.js

# Save PM2 configuration
pm2 save

# Setup PM2 startup (run as root if possible)
pm2 startup

# Check status
pm2 status
pm2 logs paynxo-api
```

**Option B: Using systemd (if you have root access)**

```bash
# Create systemd service (as root)
sudo nano /etc/systemd/system/paynxo-api.service
```

Add:
```ini
[Unit]
Description=PayNXO API Server
After=network.target

[Service]
Type=simple
User=username
Group=username
WorkingDirectory=/home/username/apps/paynxo/backend
EnvironmentFile=/home/username/apps/paynxo/backend/.env
ExecStart=/home/username/apps/paynxo/backend/paynxo-api
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

```bash
# Enable and start
sudo systemctl daemon-reload
sudo systemctl enable paynxo-api
sudo systemctl start paynxo-api
sudo systemctl status paynxo-api
```

---

### PART 5: Setup Reverse Proxy

#### Step 5.1: Using Apache (cPanel default)

Create `.htaccess` in a subdomain or path for API:

**Option A: API on subdomain (api.paynxo.com)**

1. **Create Subdomain in cPanel:**
   - Go to "Subdomains"
   - Create: `api.paynxo.com`
   - Document Root: `/home/username/public_html/api`

2. **Create .htaccess in `/home/username/public_html/api/.htaccess`:**

```apache
RewriteEngine On
RewriteCond %{REQUEST_URI} !^/\.well-known/acme-challenge/
RewriteRule ^(.*)$ http://localhost:8080/$1 [P,L]

# Enable proxy
<IfModule mod_proxy.c>
  ProxyPreserveHost On
  ProxyPass / http://localhost:8080/
  ProxyPassReverse / http://localhost:8080/
</IfModule>

# CORS Headers
<IfModule mod_headers.c>
  Header set Access-Control-Allow-Origin "https://paynxo.com"
  Header set Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS"
  Header set Access-Control-Allow-Headers "Content-Type, Authorization"
</IfModule>
```

**Option B: API on path (paynxo.com/api)**

Add to main `.htaccess` in `/home/username/public_html/.htaccess`:

```apache
# API Proxy
RewriteEngine On
RewriteCond %{REQUEST_URI} ^/api/(.*)$
RewriteRule ^api/(.*)$ http://localhost:8080/api/$1 [P,L]

# Frontend SPA routing (must come after API proxy)
RewriteCond %{REQUEST_FILENAME} !-f
RewriteCond %{REQUEST_FILENAME} !-d
RewriteRule ^ index.html [L]
```

---

### PART 6: SSL Certificate Setup

#### Step 6.1: Enable AutoSSL (Automatic)

1. **In cPanel, go to "SSL/TLS Status"**
2. Click "Run AutoSSL"
3. Wait for certificates to be issued
4. Verify: Green checkmark next to paynxo.com

#### Step 6.2: Force HTTPS

Add to top of `.htaccess` in `public_html/`:

```apache
# Force HTTPS
RewriteEngine On
RewriteCond %{HTTPS} off
RewriteRule ^(.*)$ https://%{HTTP_HOST}%{REQUEST_URI} [L,R=301]
```

---

### PART 7: Configure Frontend API Endpoint

The frontend needs to know where the backend is:

**If using subdomain (api.paynxo.com):**
- Frontend is already configured to use `https://api.paynxo.com/api`
- No changes needed!

**If using path (paynxo.com/api):**
- Frontend needs to use relative paths `/api`
- Edit frontend code before building:

```javascript
// In frontend/src/data/apiClient.js
const API_BASE_URL = '/api';  // Change from full URL to relative path
```

Then rebuild and re-upload frontend.

---

### PART 8: Testing

#### Step 8.1: Test Backend API

```bash
# From SSH
curl http://localhost:8080/api/health

# From browser (should work if proxy is configured)
https://api.paynxo.com/api/health
# OR
https://paynxo.com/api/health
```

Expected response: `{"status":"ok"}`

#### Step 8.2: Test Frontend

1. Visit: https://paynxo.com
2. Should see login page
3. Open browser console (F12)
4. Check for any errors
5. Try to login

---

### PART 9: Mobile App Build

Same as before - build on your local machine:

```bash
cd C:\paynxo\apk

# Create production environment
echo API_BASE_URL=https://api.paynxo.com/api > .env
# OR if using path-based API:
# echo API_BASE_URL=https://paynxo.com/api > .env

echo API_TIMEOUT=30000 >> .env

# Build with EAS
eas build --platform android --profile production

# Or local build
npx expo prebuild --platform android
cd android
./gradlew assembleRelease
```

---

## 🔧 cPanel-Specific Management

### File Manager Operations

**Upload Files:**
1. cPanel → File Manager
2. Navigate to target directory
3. Click "Upload" button
4. Select files or use drag-and-drop

**Extract Archives:**
1. Upload ZIP file
2. Right-click → Extract
3. Delete ZIP after extraction

**Edit Files:**
1. Right-click file → Edit
2. Make changes
3. Save

### Database Management

**phpMyAdmin:**
1. cPanel → phpMyAdmin
2. Select database
3. Import SQL files
4. Run queries

**Remote Access:**
1. cPanel → Remote MySQL
2. Add your IP address
3. Connect using MySQL client

### Monitor Backend

```bash
# Via SSH
pm2 status
pm2 logs paynxo-api
pm2 restart paynxo-api

# Or check error logs
tail -f ~/apps/paynxo/backend/logs/error.log
```

### Cron Jobs (for maintenance)

1. cPanel → Cron Jobs
2. Add restart cron (if backend crashes):

```
*/5 * * * * pm2 restart paynxo-api >/dev/null 2>&1
```

---

## 🐛 Troubleshooting

### Issue: 404 on All Routes

**Solution:** Check .htaccess SPA routing rules

### Issue: API Not Accessible

**Solution:** 
1. Check if backend is running: `pm2 status`
2. Check proxy configuration in .htaccess
3. Verify port 8080 is not blocked

### Issue: 502 Bad Gateway

**Solution:**
1. Backend not running: `pm2 restart paynxo-api`
2. Check logs: `pm2 logs paynxo-api`

### Issue: Database Connection Failed

**Solution:**
1. Verify credentials in `.env`
2. Check database exists: phpMyAdmin
3. Test connection:
```bash
mysql -h localhost -u username_paynxo -p username_paymentgateway
```

### Issue: CORS Errors

**Solution:**
1. Check ALLOWED_ORIGINS in backend `.env`
2. Restart backend: `pm2 restart paynxo-api`
3. Verify proxy headers in .htaccess

---

## 📊 Directory Structure

```
/home/username/
├── public_html/              # Frontend files
│   ├── index.html
│   ├── assets/
│   ├── .htaccess             # SPA routing + API proxy
│   └── api/                  # If using subdomain
│       └── .htaccess         # API proxy
├── apps/
│   └── paynxo/
│       ├── backend/
│       │   ├── paynxo-api    # Compiled Go binary
│       │   ├── .env          # Environment config
│       │   └── cmd/
│       └── frontend/
└── logs/
```

---

## 🔐 Security Checklist

- [x] Strong database password
- [x] JWT_SECRET changed from default
- [x] SSL certificate installed (AutoSSL)
- [x] HTTPS enforced
- [x] .env file permissions (chmod 600)
- [x] Database user has limited privileges
- [x] CORS properly configured
- [x] Security headers in .htaccess
- [ ] Regular backups configured
- [ ] Monitoring setup

---

## 📞 Common Commands

```bash
# SSH into cPanel server
ssh username@yourdomain.com

# Navigate to app
cd ~/apps/paynxo/backend

# Check backend status
pm2 status

# View logs
pm2 logs paynxo-api

# Restart backend
pm2 restart paynxo-api

# Update application
cd ~/apps/paynxo
git pull origin main
cd backend
go build -o paynxo-api ./cmd/server
pm2 restart paynxo-api

# Backup database
mysqldump -u username_paynxo -p username_paymentgateway > backup.sql
```

---

## 🎉 Deployment Complete!

Your PayNXO application should now be accessible at:
- **Frontend**: https://paynxo.com
- **API**: https://api.paynxo.com/api (or https://paynxo.com/api)

### Next Steps:

1. **Create Admin Account**
   - Login to frontend
   - Create first admin user

2. **Configure System**
   - Add payment gateways
   - Add companies
   - Add merchants

3. **Distribute Mobile App**
   - Upload APK to Google Play
   - Or share APK directly

4. **Setup Monitoring**
   - Configure uptime monitoring
   - Setup error alerts

5. **Regular Backups**
   - Setup automated database backups
   - Backup .env files

---

## 📚 Additional Resources

- **cPanel Documentation**: https://docs.cpanel.net/
- **PM2 Documentation**: https://pm2.keymetrics.io/docs/
- **GitHub Repository**: https://github.com/Dashhold/paynxo

---

**Need Help?** Check cPanel error logs:
- Apache Error Log: cPanel → Errors
- PHP Error Log: cPanel → Errors
- Backend Logs: `pm2 logs paynxo-api`
