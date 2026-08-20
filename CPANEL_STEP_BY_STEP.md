# PayNXO cPanel Deployment - Step-by-Step Walkthrough 🎬

**Follow along like a video tutorial - with screenshots descriptions**

---

## 📺 Introduction

This guide walks you through deploying PayNXO on cPanel hosting with:
- ✅ Screenshots descriptions for each step
- ✅ Exact button names and locations
- ✅ What you should see at each stage
- ✅ Troubleshooting inline

**Time Required:** 30-40 minutes

---

## 🎯 Part 1: Prepare Your Files (10 min)

### 1.1 Build Frontend on Your Computer

**Location:** Your Windows PC

**Steps:**

1. **Open Command Prompt or PowerShell**
   - Press `Windows + R`
   - Type `cmd` and press Enter

2. **Navigate to project:**
   ```bash
   cd C:\paynxo\frontend
   ```

3. **Install dependencies:**
   ```bash
   npm install
   ```
   - You'll see: Installing packages... (takes 1-2 min)

4. **Build for production:**
   ```bash
   npm run build
   ```
   - You'll see: "Building for production..."
   - Success message: "✓ built in X.XXs"
   - Creates `dist` folder

5. **Verify dist folder exists:**
   ```bash
   dir dist
   ```
   - Should see: `index.html`, `assets` folder

6. **Create ZIP file for upload:**
   - Right-click `dist` folder
   - Choose "Send to" → "Compressed (zipped) folder"
   - Name it: `frontend-dist.zip`

✅ **You should now have:** `C:\paynxo\frontend\frontend-dist.zip`

---

## 🗄️ Part 2: Setup Database in cPanel (5 min)

### 2.1 Login to cPanel

**Steps:**

1. **Open browser and go to:**
   ```
   https://yourdomain.com:2083
   ```
   - OR: `https://yourdomain.com/cpanel`

2. **Enter credentials:**
   - Username: (provided by hosting)
   - Password: (provided by hosting)
   - Click "Log in"

3. **You should see:** cPanel dashboard with icons

### 2.2 Create Database

**Steps:**

1. **Find "Databases" section** (scroll down)
   
2. **Click "MySQL Database Wizard"**
   - Icon looks like: 🗄️ cylinder/database

3. **Step 1: Create Database**
   - Database Name: `paymentgateway`
   - Click "Next Step"
   - ✅ You'll see: "Added the database username_paymentgateway"

4. **Step 2: Create User**
   - Username: `paynxo`
   - Password: Click "Password Generator"
   - Save this password! (copy to notepad)
   - Click "Create User"
   - ✅ You'll see: "Added the user username_paynxo"

5. **Step 3: Add Privileges**
   - Check "ALL PRIVILEGES" (at the top)
   - Click "Next Step"
   - ✅ You'll see: "User username_paynxo was added to database"

6. **Write down your credentials:**
   ```
   Database Name: username_paymentgateway
   Database User: username_paynxo
   Database Password: [the password you generated]
   Database Host: localhost
   ```

✅ **Database is ready!**

---

## 📁 Part 3: Upload Frontend to cPanel (10 min)

### 3.1 Open File Manager

**Steps:**

1. **In cPanel dashboard, find "Files" section**

2. **Click "File Manager"**
   - Icon looks like: 📁 folder

3. **File Manager opens in new tab**
   - You'll see folders on left side
   - Main area shows files

### 3.2 Navigate to Public HTML

**Steps:**

1. **In left sidebar, click "public_html"**
   - This is where your website files go
   - For main domain: `/public_html/`
   - For addon domain: `/public_html/yourdomain.com/`

2. **Clear existing files (if any):**
   - Select old files (if present)
   - Click "Delete" in top menu
   - Confirm deletion

### 3.3 Upload Frontend Files

**Steps:**

1. **Click "Upload" button** (top menu)
   - New page opens with upload area

2. **Click "Select File"** or drag & drop
   - Choose `C:\paynxo\frontend\frontend-dist.zip`
   - Upload starts automatically
   - ✅ Progress bar shows 100%

3. **Go back to File Manager:**
   - Click browser back button
   - You should see `frontend-dist.zip` in file list

4. **Extract ZIP file:**
   - Right-click `frontend-dist.zip`
   - Choose "Extract"
   - Click "Extract Files" button
   - ✅ You'll see: "Extraction complete"
   - Close the dialog

5. **Move files to root:**
   - Open `frontend-dist` folder (double-click)
   - Select ALL files inside (Ctrl+A or click Select All)
   - Click "Move" button (top menu)
   - In "Move to:" field, type: `/public_html/`
   - Click "Move Files"
   - ✅ Files moved

6. **Go back to public_html:**
   - Click "public_html" in left sidebar
   - You should now see:
     - `index.html` ✅
     - `assets` folder ✅
     - `frontend-dist` folder (delete this - empty)
     - `frontend-dist.zip` (delete this too)

7. **Delete empty folders:**
   - Select `frontend-dist` and `frontend-dist.zip`
   - Click "Delete"
   - Confirm

### 3.4 Create .htaccess File

**Steps:**

1. **In public_html, click "File" → "New File"**

2. **New File dialog opens:**
   - New File Name: `.htaccess`
   - Click "Create New File"
   - ✅ You'll see: "File created successfully"

3. **Right-click `.htaccess` → "Edit"**

4. **Paste this code:**

```apache
# Force HTTPS
RewriteEngine On
RewriteCond %{HTTPS} off
RewriteRule ^(.*)$ https://%{HTTP_HOST}%{REQUEST_URI} [L,R=301]

# SPA Routing
RewriteCond %{REQUEST_FILENAME} !-f
RewriteCond %{REQUEST_FILENAME} !-d
RewriteRule ^ index.html [L]

# Gzip Compression
<IfModule mod_deflate.c>
  AddOutputFilterByType DEFLATE text/html text/plain text/xml text/css text/javascript application/javascript application/json
</IfModule>

# Cache Static Assets
<IfModule mod_expires.c>
  ExpiresActive On
  ExpiresByType image/jpg "access plus 1 year"
  ExpiresByType image/jpeg "access plus 1 year"
  ExpiresByType image/gif "access plus 1 year"
  ExpiresByType image/png "access plus 1 year"
  ExpiresByType text/css "access plus 1 year"
  ExpiresByType application/javascript "access plus 1 year"
</IfModule>

# Security Headers
<IfModule mod_headers.c>
  Header set X-Content-Type-Options "nosniff"
  Header set X-Frame-Options "SAMEORIGIN"
  Header set X-XSS-Protection "1; mode=block"
</IfModule>
```

5. **Click "Save Changes"** (top right)
   - ✅ You'll see: "File saved"

6. **Close editor**

✅ **Frontend is uploaded! Test it:** `http://yourdomain.com`
- Should see login page (might show errors for API - that's OK for now)

---

## 🔧 Part 4: Deploy Backend via SSH (15 min)

### 4.1 Connect to SSH

**Note:** You need SSH access. If you don't have it, contact your hosting provider.

**Windows (Using PuTTY or Terminal):**

**Option A: Windows Terminal (Windows 10/11):**

1. **Open Windows Terminal or Command Prompt**

2. **Connect:**
   ```bash
   ssh username@yourdomain.com
   ```
   - Replace `username` with your cPanel username
   - Replace `yourdomain.com` with your domain or server IP

3. **Enter password** when prompted
   - Type your cPanel password (won't show characters)
   - Press Enter

4. **✅ You should see:** Command prompt like `[username@server ~]$`

**Option B: cPanel Terminal (if available):**

1. **In cPanel, search for "Terminal"**
2. **Click "Terminal"** icon
3. **Terminal opens in browser**
4. **✅ You're in!**

### 4.2 Check/Install Go

**Steps:**

1. **Check if Go is installed:**
   ```bash
   go version
   ```

2. **If you see version (e.g., "go1.21.6"):**
   - ✅ Go is installed, skip to step 4.3

3. **If "command not found", install Go:**
   ```bash
   cd ~
   wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
   tar -xzf go1.21.6.linux-amd64.tar.gz
   echo 'export PATH=$HOME/go/bin:$PATH' >> ~/.bashrc
   source ~/.bashrc
   go version
   ```
   - ✅ Should now show: "go version go1.21.6"

### 4.3 Clone Repository

**Steps:**

1. **Create app directory:**
   ```bash
   mkdir -p ~/apps/paynxo
   cd ~/apps/paynxo
   ```

2. **Clone from GitHub:**
   ```bash
   git clone https://github.com/Dashhold/paynxo.git .
   ```
   - You'll see: "Cloning into..."
   - ✅ Success: "Resolving deltas: 100%"

3. **Verify files:**
   ```bash
   ls -la
   ```
   - Should see: `backend/`, `frontend/`, `apk/`

### 4.4 Configure Backend

**Steps:**

1. **Navigate to backend:**
   ```bash
   cd ~/apps/paynxo/backend
   ```

2. **Create .env file:**
   ```bash
   nano .env
   ```
   - Terminal text editor opens

3. **Paste this configuration:**
   ```env
   PORT=8080
   GIN_MODE=release

   DB_HOST=localhost
   DB_PORT=3306
   DB_NAME=username_paymentgateway
   DB_USER=username_paynxo
   DB_PASSWORD=your_database_password_here

   JWT_SECRET=CHANGE_THIS_TO_VERY_LONG_RANDOM_STRING
   JWT_EXPIRY=24h

   ALLOWED_ORIGINS=https://paynxo.com,https://www.paynxo.com

   APP_ENV=production
   ```

4. **IMPORTANT - Update these values:**
   - `DB_NAME`: Your database name from Part 2
   - `DB_USER`: Your database user from Part 2
   - `DB_PASSWORD`: Your database password from Part 2
   - `JWT_SECRET`: Generate random string:
     ```bash
     # Press Ctrl+X to exit nano (don't save yet)
     # Generate secret:
     openssl rand -base64 64
     # Copy the output
     # Re-open nano:
     nano .env
     # Paste the secret in JWT_SECRET line
     ```

5. **Save file:**
   - Press `Ctrl + X`
   - Press `Y` (Yes to save)
   - Press `Enter` (confirm filename)
   - ✅ You're back to command prompt

6. **Verify .env was created:**
   ```bash
   cat .env
   ```
   - Should show your configuration

### 4.5 Build Backend

**Steps:**

1. **Build the Go application:**
   ```bash
   cd ~/apps/paynxo/backend
   go build -o paynxo-api ./cmd/server
   ```
   - This takes 1-2 minutes
   - You'll see: "Downloading packages..."
   - ✅ Success: No errors, back to prompt

2. **Verify binary was created:**
   ```bash
   ls -lh paynxo-api
   ```
   - Should show file size (e.g., "20M")

3. **Test run (optional):**
   ```bash
   ./paynxo-api
   ```
   - Should start server
   - ✅ You'll see logs scrolling
   - Press `Ctrl + C` to stop

### 4.6 Setup PM2 Process Manager

**Steps:**

1. **Install PM2:**
   ```bash
   npm install -g pm2
   ```
   - Takes 30 seconds
   - ✅ You'll see: "added 1 package"

2. **Start backend with PM2:**
   ```bash
   cd ~/apps/paynxo/backend
   pm2 start paynxo-api --name paynxo-api
   ```
   - ✅ You'll see table showing:
     - name: paynxo-api
     - status: online ✅
     - uptime: 0s

3. **Save PM2 configuration:**
   ```bash
   pm2 save
   ```
   - ✅ You'll see: "Successfully saved"

4. **Setup auto-restart on reboot:**
   ```bash
   pm2 startup
   ```
   - Follow the command it shows (copy and paste)
   - ✅ You'll see: "Startup successfully added"

5. **Check status:**
   ```bash
   pm2 status
   ```
   - Should show paynxo-api as "online"

6. **View logs:**
   ```bash
   pm2 logs paynxo-api --lines 20
   ```
   - Should see server logs
   - Press `Ctrl + C` to exit

✅ **Backend is running!**

---

## 🔗 Part 5: Setup API Proxy (5 min)

### 5.1 Create API Subdomain

**Steps:**

1. **Go back to cPanel dashboard**

2. **Find "Domains" section**

3. **Click "Subdomains"**

4. **Create subdomain:**
   - Subdomain: `api`
   - Domain: Select `paynxo.com` from dropdown
   - Document Root: Will auto-fill (e.g., `/public_html/api`)
   - Click "Create"
   - ✅ You'll see: "Success: Subdomain created"

### 5.2 Configure API Proxy

**Steps:**

1. **Go to File Manager**

2. **Navigate to:** `/public_html/api/`

3. **Create .htaccess file:**
   - Click "File" → "New File"
   - Name: `.htaccess`
   - Click "Create New File"

4. **Edit .htaccess:**
   - Right-click `.htaccess` → "Edit"
   - Paste this code:

```apache
RewriteEngine On

# Proxy all requests to backend
RewriteRule ^(.*)$ http://localhost:8080/$1 [P,L]

# CORS Headers
<IfModule mod_headers.c>
  Header always set Access-Control-Allow-Origin "https://paynxo.com"
  Header always set Access-Control-Allow-Methods "GET, POST, PUT, DELETE, OPTIONS"
  Header always set Access-Control-Allow-Headers "Content-Type, Authorization"
  Header always set Access-Control-Max-Age "3600"
</IfModule>

# Handle OPTIONS preflight
<IfModule mod_rewrite.c>
  RewriteCond %{REQUEST_METHOD} OPTIONS
  RewriteRule ^(.*)$ $1 [R=200,L]
</IfModule>
```

5. **Save file** (top right button)

6. **Close editor**

✅ **API proxy is configured!**

---

## 🔒 Part 6: Enable SSL (2 min)

### 6.1 Install SSL Certificate

**Steps:**

1. **In cPanel, search for "SSL"**

2. **Click "SSL/TLS Status"**

3. **Find your domains in list:**
   - `paynxo.com`
   - `www.paynxo.com`
   - `api.paynxo.com`

4. **Click "Run AutoSSL"** (button at top)

5. **Wait 1-2 minutes**
   - You'll see progress
   - ✅ Green checkmarks appear next to domains

6. **Verify:**
   - All three domains should show: ✅ Certificate installed

✅ **SSL is enabled!**

---

## ✅ Part 7: Test Everything (5 min)

### 7.1 Test Frontend

**Steps:**

1. **Open browser**

2. **Go to:** `https://paynxo.com`

3. **✅ You should see:** Login page

4. **Check console (F12):**
   - No red errors (some warnings OK)

### 7.2 Test API

**Steps:**

1. **In browser, open new tab**

2. **Go to:** `https://api.paynxo.com/api/health`

3. **✅ You should see:** 
   ```json
   {"status":"ok"}
   ```

4. **If you see error:**
   - Check PM2 status: `pm2 status`
   - Check logs: `pm2 logs paynxo-api`
   - Restart: `pm2 restart paynxo-api`

### 7.3 Test Login

**Steps:**

1. **Go to:** `https://paynxo.com`

2. **Try to login** with test credentials

3. **✅ Success if:**
   - No CORS errors in console
   - Login either succeeds or shows proper error

4. **If CORS error:**
   - Check ALLOWED_ORIGINS in backend `.env`
   - Should include: `https://paynxo.com`
   - Restart backend: `pm2 restart paynxo-api`

✅ **Everything is working!**

---

## 🎉 Deployment Complete!

### Your application is live:
- **Frontend**: https://paynxo.com
- **API**: https://api.paynxo.com

### What's Next?

1. **Create Admin Account**
2. **Configure Payment Gateways**
3. **Build Mobile App** (see CPANEL_QUICK_START.md)
4. **Setup Monitoring**
5. **Regular Backups**

---

## 🆘 Troubleshooting

### "API is not accessible"
```bash
# Check if backend is running
pm2 status

# Should show: online
# If not online:
pm2 restart paynxo-api
pm2 logs paynxo-api
```

### "Database connection failed"
```bash
# Test database connection
mysql -h localhost -u username_paynxo -p username_paymentgateway

# If fails, check:
# 1. Database credentials in .env
# 2. Database exists in cPanel
# 3. User has privileges
```

### "CORS Error"
```bash
# Check ALLOWED_ORIGINS in .env
cat ~/apps/paynxo/backend/.env | grep ALLOWED_ORIGINS

# Should include your domain
# Update if needed:
nano ~/apps/paynxo/backend/.env

# Restart:
pm2 restart paynxo-api
```

---

## 📞 Need More Help?

**Read these guides:**
- `CPANEL_DEPLOYMENT_GUIDE.md` - Detailed guide
- `CPANEL_QUICK_START.md` - Quick reference

**Check logs:**
```bash
# Backend logs
pm2 logs paynxo-api

# Apache error logs (in cPanel → Errors)
```

---

**Congratulations! Your PayNXO application is deployed! 🚀**
