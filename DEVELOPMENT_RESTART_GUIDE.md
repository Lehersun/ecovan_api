# Development Restart Guide

## 🚀 How to Restart Application After Code Changes

### **After ANY code changes, you MUST restart the application:**

```bash
# 1. Stop the development environment
make dev-stop

# 2. Build the application (optional but recommended)
make build

# 3. Start the development environment fresh
make dev
```

### **Why Restart is Required:**
- Go applications compile and run the code at startup
- **Code changes are NOT hot-reloaded**
- Old compiled code continues running until restart
- Changes only take effect after complete restart

---

## 🔧 How to Add New Endpoints Properly

### **Step 1: Update the API Endpoints List**
**File:** `internal/app/server.go` (around line 105)

**Find this section:**
```go
r.Get("/", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, `{"message":"API v1","endpoints":["/healthz","/metrics","/auth/login","/auth/refresh","/users","/clients","/warehouses","/equipment","/drivers"]}`)
})
```

**Add your new endpoint to the list:**
```go
fmt.Fprintf(w, `{"message":"API v1","endpoints":["/healthz","/metrics","/auth/login","/auth/refresh","/users","/clients","/warehouses","/equipment","/drivers","/your-new-endpoint"]}`)
```

### **Step 2: Implement the Endpoint Logic**
**File:** `internal/app/server.go` (after line 290)

**Add your route:**
```go
// Protected your-endpoint management
r.Route("/your-endpoint", func(r chi.Router) {
    // Create handler and middleware
    yourRepo := pg.NewYourRepository(db.GetPool())
    yourService := service.NewYourService(yourRepo)
    yourHandler := httpmiddleware.NewYourHandler(yourService)
    yourJWTManager := auth.NewDefaultJWTManager(cfg.Auth.JWTSecret)
    authMiddleware := httpmiddleware.NewAuthMiddleware(yourJWTManager)
    rbacMiddleware := httpmiddleware.NewRBACMiddleware()

    // Require authentication
    r.Use(authMiddleware.RequireAuth)

    // Read endpoints - accessible by all authenticated users
    r.With(rbacMiddleware.RequireReadAccess).Group(func(r chi.Router) {
        r.Get("/", yourHandler.ListItems)
        r.Get("/{id}", yourHandler.GetItem)
    })

    // Write endpoints - ADMIN and DISPATCHER only
    r.With(rbacMiddleware.RequireWriteAccess).Group(func(r chi.Router) {
        r.Post("/", yourHandler.CreateItem)
        r.Put("/{id}", yourHandler.UpdateItem)
        r.Delete("/{id}", yourHandler.DeleteItem)
        r.Post("/{id}/restore", yourHandler.RestoreItem)
    })
})
```

### **Step 3: Create Required Files**
- **Handler:** `internal/adapter/http/your_handler.go`
- **Service:** `internal/service/your_service.go`
- **Repository:** `internal/adapter/repo/pg/your_repo.go`
- **Models:** `internal/models/your.go`
- **Ports:** `internal/port/your_repo.go` and `internal/port/your_service.go`

### **Step 4: Restart Application**
```bash
make dev-stop
make build
make dev
```

---

## ⚠️ Common Mistakes to Avoid

### **❌ Don't Do This:**
- Make code changes and expect them to work immediately
- Only restart the database (`make db`) - this doesn't restart the Go app
- Forget to update the endpoints list in the API info
- Skip the build step after major changes

### **✅ Always Do This:**
- Restart the entire development environment after code changes
- Update the hardcoded endpoints list
- Test endpoints after restart
- Use `make dev` for complete restart

---

## 🔍 Verification Checklist

After adding endpoints and restarting:

1. **Check API endpoints list:**
   ```bash
   curl -s http://localhost:8080/api/v1/ | jq .
   ```

2. **Test your new endpoint:**
   ```bash
   # Get admin token first
   TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{"email":"admin@example.com","password":"admin123456"}' | \
     grep -o '"accessToken":"[^"]*"' | cut -d'"' -f4)
   
   # Test your endpoint
   curl -s -H "Authorization: Bearer $TOKEN" \
     http://localhost:8080/api/v1/your-endpoint | jq .
   ```

3. **Verify it appears in the endpoints list**

---

## 📝 Quick Reference

```bash
# Development workflow:
# 1. Make code changes
# 2. make dev-stop
# 3. make build (optional)
# 4. make dev
# 5. Test endpoints

# Remember: Code changes require FULL restart!
```
