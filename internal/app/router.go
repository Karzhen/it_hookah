package app

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/karzhen/restaurant-lk/internal/handler"
	"github.com/karzhen/restaurant-lk/internal/middleware"
	"github.com/karzhen/restaurant-lk/internal/utils/apperror"
	"github.com/karzhen/restaurant-lk/internal/utils/httpx"
)

type RouterDependencies struct {
	AuthHandler         *handler.AuthHandler
	UserHandler         *handler.UserHandler
	HealthHandler       *handler.HealthHandler
	CartHandler         *handler.CartHandler
	OrderHandler        *handler.OrderHandler
	MixHandler          *handler.MixHandler
	TagHandler          *handler.TagHandler
	StockMovementHandle *handler.StockMovementHandler
	CatalogPublicHandle *handler.CatalogPublicHandler
	CatalogAdminHandle  *handler.CatalogAdminHandler
	AuthMW              *middleware.AuthMiddleware
	Logger              *slog.Logger
}

func NewRouter(deps RouterDependencies) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestIDMiddleware)
	r.Use(middleware.RecoveryMiddleware(deps.Logger))
	r.Use(middleware.LoggingMiddleware(deps.Logger))
	r.Use(middleware.CORSMiddleware())

	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteError(deps.Logger, w, apperror.New("not_found", "route not found", http.StatusNotFound))
	})

	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		httpx.WriteError(deps.Logger, w, apperror.New("method_not_allowed", "method not allowed", http.StatusMethodNotAllowed))
	})

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/health", deps.HealthHandler.Health)
		api.Get("/categories", deps.CatalogPublicHandle.ListCategories)
		api.Get("/tobacco/flavors", deps.CatalogPublicHandle.ListFlavors)
		api.Get("/tobacco/strengths", deps.CatalogPublicHandle.ListStrengths)
		api.Get("/products", deps.CatalogPublicHandle.ListProducts)
		api.Get("/products/{id}", deps.CatalogPublicHandle.GetProductByID)
		api.Get("/tags", deps.TagHandler.ListPublicTags)

		api.Route("/auth", func(auth chi.Router) {
			auth.Post("/register", deps.AuthHandler.Register)
			auth.Post("/login", deps.AuthHandler.Login)
			auth.Post("/refresh", deps.AuthHandler.Refresh)
			auth.Post("/logout", deps.AuthHandler.Logout)

			auth.Group(func(protected chi.Router) {
				protected.Use(deps.AuthMW.RequireAuth)
				protected.Patch("/change-password", deps.AuthHandler.ChangePassword)
			})
		})

		api.Route("/users", func(users chi.Router) {
			users.Use(deps.AuthMW.RequireAuth)
			users.Get("/me", deps.UserHandler.Me)
			users.Patch("/me", deps.UserHandler.UpdateMe)
		})

		api.Group(func(cart chi.Router) {
			cart.Use(deps.AuthMW.RequireAuth)
			cart.Get("/cart", deps.CartHandler.GetCart)
			cart.Post("/cart/items", deps.CartHandler.AddItem)
			cart.Patch("/cart/items/{id}", deps.CartHandler.UpdateItemQuantity)
			cart.Delete("/cart/items/{id}", deps.CartHandler.RemoveItem)
			cart.Delete("/cart/items", deps.CartHandler.ClearCart)
		})

		api.Group(func(orders chi.Router) {
			orders.Use(deps.AuthMW.RequireAuth)
			orders.Post("/orders", deps.OrderHandler.CreateOrder)
			orders.Get("/orders", deps.OrderHandler.ListMyOrders)
			orders.Get("/orders/{id}", deps.OrderHandler.GetMyOrderByID)
		})

		api.Group(func(mixes chi.Router) {
			mixes.Use(deps.AuthMW.RequireAuth)
			mixes.Get("/mixes", deps.MixHandler.ListPublicMixes)
			mixes.Get("/mixes/{id}", deps.MixHandler.GetPublicMixByID)
		})

		api.Route("/admin", func(admin chi.Router) {
			admin.Use(deps.AuthMW.RequireAuth)
			admin.Use(deps.AuthMW.RequireRole("admin"))

			admin.Get("/categories", deps.CatalogAdminHandle.ListCategories)
			admin.Post("/categories", deps.CatalogAdminHandle.CreateCategory)
			admin.Patch("/categories/{id}", deps.CatalogAdminHandle.UpdateCategory)
			admin.Delete("/categories/{id}", deps.CatalogAdminHandle.DeleteCategory)

			admin.Get("/tobacco/flavors", deps.CatalogAdminHandle.ListFlavors)
			admin.Post("/tobacco/flavors", deps.CatalogAdminHandle.CreateFlavor)
			admin.Patch("/tobacco/flavors/{id}", deps.CatalogAdminHandle.UpdateFlavor)
			admin.Delete("/tobacco/flavors/{id}", deps.CatalogAdminHandle.DeleteFlavor)

			admin.Get("/tobacco/strengths", deps.CatalogAdminHandle.ListStrengths)
			admin.Post("/tobacco/strengths", deps.CatalogAdminHandle.CreateStrength)
			admin.Patch("/tobacco/strengths/{id}", deps.CatalogAdminHandle.UpdateStrength)
			admin.Delete("/tobacco/strengths/{id}", deps.CatalogAdminHandle.DeleteStrength)

			admin.Get("/products", deps.CatalogAdminHandle.ListProducts)
			admin.Post("/products", deps.CatalogAdminHandle.CreateProduct)
			admin.Patch("/products/{id}", deps.CatalogAdminHandle.UpdateProduct)
			admin.Delete("/products/{id}", deps.CatalogAdminHandle.DeleteProduct)
			admin.Patch("/products/{id}/stock", deps.CatalogAdminHandle.UpdateProductStock)
			admin.Get("/products/{id}/stock-movements", deps.StockMovementHandle.ListProductStockMovements)
			admin.Get("/stock-movements", deps.StockMovementHandle.ListStockMovements)

			admin.Get("/orders", deps.OrderHandler.ListAllOrders)
			admin.Patch("/orders/{id}/status", deps.OrderHandler.UpdateOrderStatus)

			admin.Get("/mixes", deps.MixHandler.ListAdminMixes)
			admin.Post("/mixes", deps.MixHandler.CreateMix)
			admin.Patch("/mixes/{id}", deps.MixHandler.UpdateMix)
			admin.Delete("/mixes/{id}", deps.MixHandler.DeleteMix)

			admin.Get("/tags", deps.TagHandler.ListAdminTags)
			admin.Post("/tags", deps.TagHandler.CreateTag)
			admin.Patch("/tags/{id}", deps.TagHandler.UpdateTag)
			admin.Delete("/tags/{id}", deps.TagHandler.DeleteTag)
		})
	})

	return r
}
