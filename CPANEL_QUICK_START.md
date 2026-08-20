# PayNXO cPanel - Quick Start Guide ⚡

**Ultra-fast deployment for cPanel hosting**

---

## ⚠️ Before You Start

**Check if your cPanel hosting supports:**
- ✅ SSH Access (Required for backend)
- ✅ Node.js (For PM2)
- ✅ MySQL Database
- ✅ SSL Certificates

**If NO SSH access:** You can only deploy the frontend. Backend needs VPS/Dedicated hosting.

---

## 🚀 5-Step Deployment

### STEP 1: Setup Database (5 min)

**In cPanel:**

1. Go to **MySQL Database Wizard**
2. Create Database: `paymentgateway`
3. Create User: `paynxo` with strong password
4. Grant ALL PRIVILEGES
5. **Save these credentials!**

---

### STEP 2: Deploy Frontend (10 min)

**On Your Computer:**

```bash
cd C:\paynxo\frontend
npm install
npm run build
```

**In cPanel File Manager:**

1. Go to `public_html/`
2. Upload ALL files from `C:\paynxo\frontend\dist\`
3. Create `.htaccess` file:

```apache
RewriteEngine On
RewriteBase /

# SPA Routing
RewriteCond %{REQUEST_FILENAME} !-f
RewriteCond %{REQUEST_FILENAME} !-d
RewriteRule ^ index.html [L]

# Force HTTPS
RewriteCond %{HTTPS} off
RewriteRule ^(.*)$ https://%{HTTP_HOST}%{REQUEST_URI} [L,R=301]

# Gzip
<IfModule mod_deflate.c>
  AddOutputFilterByType DEFLATE text/html text/css application/javascript
</IfModule>
```

---

### STEP 3: Deploy Backend (15 min)

**Connect via SSH:**

```bash
ssh username@yourdomain.com
```

**Clone & Build:**

```bash
# Create directory
mkdir -p ~/apps/paynxo
cd ~/apps/paynxo

# Clone repository
git clone https://github.com/Dashhold/paynxo.git .

# Build backend
cd backend
go build -o paynxo-api ./cmd/server
```

**Configure:**

```bash
cd ~/apps/paynxo/backend
nano .env
```

Add:
```env
PORT=8080
GIN_MODE=release

DB_HOST=localhost
DB_PORT=3306
DB_NAME=username_paymentgateway
DB_USER=username_paynxo
DB_PASSWORD=YOUR_DB_PASSWORD

JWT_SECRET=CHANGE_THIS_TO_RANDOM_STRING
JWT_EXPIRY=24h

ALLOWED_ORIGINS=https://paynxo.com,https://www.paynxo.com

APP_ENV=production
```

**Start Backend:**

```bash
# Install PM2
npm install -g pm2

# Start
pm2 start paynxo-api --name paynxo-api

# Save
pm2 save
pm2 startup
```

---

### STEP 4: Setup API Proxy (5 min)

**Option A: Subdomain (Recommended)**

1. **In cPanel → Subdomains:**
   - Create: `api.paynxo.com`

2. **Create `.htaccess` in `/public_html/api/.htaccess`:**

```apache
RewriteEngine On
RewriteRule ^(.*)$ http://localhost:8080/$1 [P,L]

<IfModule mod_headers.c>
  Header set Access-Control-Allow-Origin "https://paynxo.com"
</IfModule>
```

**Option B: Path (paynxo.com/api)**

Edit main `.htaccess`:

```apache
# Add BEFORE SPA routing
RewriteEngine On
RewriteCond %{REQUEST_URI} ^/api/(.*)$
RewriteRule ^api/(.*)$ http://localhost:8080/api/$1 [P,L]

# Then SPA routing...
```

---

### STEP 5: Enable SSL (2 min)

**In cPanel:**

1. Go to **SSL/TLS Status**
2. Click **Run AutoSSL**
3. Wait for green checkmark

**Done!** ✅

---

## 🧪 Test Your Deployment

### Test Frontend:
```
https://paynxo.com
```
Should show login page

### Test API:
```bash
curl https://api.paynxo.com/api/health
# OR
curl https://paynxo.com/api/health
```
Should return: `{"status":"ok"}`

### Test Login:
1. Go to https://paynxo.com
2. Try logging in
3. Check browser console (F12) for errors

---

## 📱 Mobile App Build

**On your computer:**

```bash
cd C:\paynxo\apk

# Set API URL
echo API_BASE_URL=https://api.paynxo.com/api > .env
echo API_TIMEOUT=30000 >> .env

# Build
npm install -g eas-cli
eas login
eas build --platform android --profile production
```

Download APK from Expo dashboard.

---

## 🔧 Management Commands

```bash
# Connect to server
ssh username@yourdomain.com

# Check backend status
pm2 status

# View logs
pm2 logs paynxo-api

# Restart backend
pm2 restart paynxo-api

# Update app
cd ~/apps/paynxo
git pull origin main
cd backend
go build -o paynxo-api ./cmd/server
pm2 restart paynxo-api
```

---

## 🐛 Quick Fixes

### Backend Not Running?
```bash
pm2 restart paynxo-api
pm2 logs paynxo-api
```

### API 404 Error?
Check proxy .htaccess and verify backend is running

### CORS Error?
Update ALLOWED_ORIGINS in .env and restart backend

### Database Connection Error?
Verify database credentials in .env match cPanel database

---

## 📂 File Locations

```
/home/username/
├── public_html/           # Frontend
│   ├── index.html
│   ├── assets/
│   └── .htaccess
├── public_html/api/       # API proxy (if subdomain)
│   └── .htaccess
└── apps/paynxo/backend/   # Backend
    ├── paynxo-api
    └── .env
```

---

## ✅ Checklist

- [ ] Database created in cPanel
- [ ] Frontend uploaded to public_html
- [ ] .htaccess configured for SPA routing
- [ ] Backend built and running (pm2)
- [ ] API proxy configured
- [ ] SSL certificate installed
- [ ] Can access https://paynxo.com
- [ ] Can access API endpoint
- [ ] Can login successfully

---

## 🎉 You're Done!

**Your app is live at:**
- Frontend: https://paynxo.com
- API: https://api.paynxo.com/api

**Next:**
1. Create admin account
2. Add payment gateways
3. Distribute mobile APK

---

## 📞 Need Help?

Read full guide: `CPANEL_DEPLOYMENT_GUIDE.md`

**Common Issues:**
- No SSH? Contact your hosting provider
- Backend won't start? Check Go is installed
- 502 Error? Backend not running (pm2 restart)
