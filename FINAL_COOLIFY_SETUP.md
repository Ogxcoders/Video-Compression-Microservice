# FINAL COOLIFY SETUP - Complete Fix

## Current Status
✅ Docker Compose mode is working (3 containers are running)  
❌ Environment variables not being passed to containers  
❌ Domain api.trendss.net not configured  

---

## CRITICAL FIX REQUIRED

The logs show:
- **PostgreSQL failing**: `POSTGRES_PASSWORD is not specified`
- **App failing**: `API_KEY is not set`, `DATABASE_URL is required`

This happens because Coolify environment variables aren't being picked up.

---

## SOLUTION - 3 STEPS

### STEP 1: Add Environment Variables in Coolify UI

**IMPORTANT:** Add these EXACTLY as shown (name=value format):

```
API_KEY=sk_test_4f9b2c8a1e6d3f7a9b2c8e1d6f3a7b9c2e8d1f6a3b7c9e2d8f1a6b3c7e9d2f8a1
ALLOWED_DOMAINS=https://capcut.ogtemplate.com/,https://ogtemplate.com/
DATABASE_URL=postgres://compressor:compressor_secure_pw_9x8c7v6b5n4m3@db:5432/compression?sslmode=disable
WORDPRESS_API_URL=https://capcut.ogtemplate.com/wp-json/wp/v2
WORDPRESS_USERNAME=vps
WORDPRESS_APP_PASSWORD=bisf lAxw AsTk Jm2t ytUb 3ENg
POSTGRES_PASSWORD=compressor_secure_pw_9x8c7v6b5n4m3
```

**Where to add them:**
1. Go to your application in Coolify
2. Click **"Environment Variables"** tab
3. For EACH variable above:
   - Click **"+ Add"**
   - Enter the **Key** (e.g., `API_KEY`)
   - Enter the **Value** (e.g., `sk_test_4f9b...`)
   - **DO NOT** check "Build Variable?" checkbox
   - Click **Save**

---

### STEP 2: Configure Domain in Coolify

**For the domain `api.trendss.net` to work:**

#### A. DNS Setup (Do this FIRST)
1. Go to your domain registrar (GoDaddy, Namecheap, Cloudflare, etc.)
2. Add an **A Record**:
   - **Name/Host**: `api`
   - **Type**: `A`
   - **Value/Points to**: `YOUR_COOLIFY_SERVER_IP`
   - **TTL**: `Auto` or `300`

#### B. Coolify Domain Configuration
1. In Coolify, go to your application
2. Find **"Domains"** or **"Settings"** section
3. In the domain field, enter: `https://api.trendss.net`
   - ⚠️ **MUST use `https://`** prefix!
   - ⚠️ **NOT** `http://` or just `api.trendss.net`
4. Set **Port** to: `3000`
5. Save

#### C. Wait for SSL
- Coolify will automatically request SSL certificate from Let's Encrypt
- This takes 1-2 minutes
- Check proxy logs if it fails

---

### STEP 3: Verify Firewall & Proxy

**Make sure these ports are open:**
- Port **80** (for Let's Encrypt verification)
- Port **443** (for HTTPS traffic)
- Port **3000** (for your app)

**Check Coolify proxy is running:**
1. Go to **Dashboard → Servers → Your Server**
2. Click **"Proxy"** tab
3. If it says "Stopped", click **"Start Proxy"**

---

## VERIFICATION STEPS

After deploying with the changes:

### 1. Check Container Logs
In Coolify, check logs for each service:

**App logs should show:**
```
✓ Connected to PostgreSQL database
✓ Server starting on port 3000
```

**PostgreSQL logs should show:**
```
✓ database system is ready to accept connections
```

**Redis logs should show:**
```
✓ Ready to accept connections tcp
```

### 2. Test Your API
```bash
# Test health endpoint
curl https://api.trendss.net/health

# Should return successful response
```

---

## TROUBLESHOOTING

### If environment variables still don't work:

**Option 1: Check if variables are being read**
1. In Coolify, go to your app
2. Click the **"Show Deployable Compose"** button
3. Verify that `${API_KEY}` is replaced with actual value
4. If not, variables aren't being substituted

**Option 2: Restart the deployment**
1. Click **"Force Rebuild"**
2. Wait for all 3 containers to restart

### If domain still doesn't work:

**Check DNS propagation:**
```bash
nslookup api.trendss.net
# Should return your server IP
```

**Check SSL certificate:**
- Go to **Dashboard → Servers → Proxy → Logs**
- Look for Let's Encrypt certificate generation messages
- If you see errors, port 80 might be closed

**Common domain issues:**
- Domain entered without `https://` → Add `https://`
- DNS not pointed to server → Update A record
- Proxy not running → Start it in Dashboard
- Port 80/443 blocked → Open in firewall

---

## EXPECTED FINAL STATE

✅ **3 containers running:**
- `app` (compressor-api)
- `db` (compressor-db) 
- `redis` (compressor-redis)

✅ **No error logs:**
- No "API_KEY is not set"
- No "DATABASE_URL is required"
- No "POSTGRES_PASSWORD" errors

✅ **Domain accessible:**
- `https://api.trendss.net` loads
- `/health` endpoint responds

✅ **SSL certificate:**
- Green padlock in browser
- Valid Let's Encrypt certificate

---

## QUICK CHECKLIST

Before redeploying, verify:

- [ ] All 7 environment variables added in Coolify UI
- [ ] "Build Variable?" checkbox is **NOT** checked for any variable
- [ ] DNS A record points `api` to your server IP
- [ ] Domain in Coolify is `https://api.trendss.net` (with https)
- [ ] Port set to `3000`
- [ ] Coolify proxy is running
- [ ] Ports 80 and 443 are open in firewall
- [ ] Docker Compose file path is `docker-compose.coolify.yml`

Then click **"Redeploy"** and wait 2-3 minutes.

---

Your API will then be live at: **https://api.trendss.net** 🚀
