package payment

import "github.com/gofiber/fiber/v2"

type Router struct {
	handler *Handler
}

func NewRouter(handler *Handler) *Router {
	return &Router{handler: handler}
}

func (r *Router) Register(app *fiber.App) {
	if r == nil || r.handler == nil || app == nil {
		return
	}

	group := app.Group("/payment")
	group.Get("/balance", r.handler.CheckBalance)
	group.Post("/invoice", r.handler.CreateInvoice)
	group.Get("/invoice/:id", r.handler.GetInvoiceById)
	group.Post("/invoice/:id/expire", r.handler.ExpireInvoice)
	group.Post("/webhook/xendit/invoice", r.handler.InvoiceWebhook)
}
