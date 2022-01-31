package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func (app *application) routes() *chi.Mux {
	r := chi.NewRouter()
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Requested-With", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
	r.Use(cors.Handler)
	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// r.Use(app.getQueryParams)

	// RESTy routes for "invoices" resource
	r.Route("/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Post("/users", app.registerUserHandler)
			r.Post("/auth", app.loginHandler)
		})

		r.Group(func(r chi.Router) {
			r.Use(app.authenticate)
			r.Get("/auth/user", app.showUserHandler)
		})

		r.Route("/organisations", func(r chi.Router) {
			r.Use(app.authenticate)
			{
				r.Get("/", app.listOrganisationsHandler)
				r.Get("/{organisationID}", app.showOrganisationHandler)
				r.Post("/", app.createOrganisationHandler)
				r.Patch("/{organisationID}", app.updateOrganisationHandler)
				r.Delete("/{organisationID}", app.deleteOrganisationHandler)

				r.Get("/{organisationID}/bank_accounts", app.listBankAccountsHandler)
				r.Get("/{organisationID}/bank_accounts/{ID}", app.showBankAccountHandler)
				r.Post("/{organisationID}/bank_accounts", app.createBankAccountHandler)
				r.Patch("/{organisationID}/bank_accounts/{ID}", app.updateBankAccountHandler)
				r.Delete("/{organisationID}/bank_accounts/{ID}", app.deleteBankAccountHandler)
			}
		})

		r.Route("/companies", func(r chi.Router) {
			r.Use(app.authenticate)
			{
				r.Get("/", app.listCompaniesHandler)
				r.Get("/search", app.searchCompaniesHandler)
				r.Get("/{companyID}", app.showCompanyHandler)
				r.Post("/", app.createCompanyHandler)
				r.Patch("/{companyID}", app.updateCompanyHandler)
				r.Delete("/{companyID}", app.deleteCompanyHandler)

				r.Get("/{companyID}/contacts", app.listContactsHandler)
				r.Get("/{companyID}/contacts/{ID}", app.showContactHandler)
				r.Post("/{companyID}/contacts", app.createContactHandler)
				r.Patch("/{companyID}/contacts/{ID}", app.updateContactHandler)
				r.Delete("/{companyID}/contacts/{ID}", app.deleteContactHandler)
			}
		})

		r.Route("/agreements", func(r chi.Router) {
			r.Use(app.authenticate)
			{
				r.Get("/", app.listAgreementsHandler)
				r.Get("/{agreementID}", app.showAgreementHandler)
				r.Post("/", app.createAgreementHandler)
				r.Patch("/{agreementID}", app.updateAgreementHandler)
				r.Delete("/{agreementID}", app.deleteAgreementHandler)
			}
		})

		r.Route("/projects", func(r chi.Router) {
			r.Use(app.authenticate)
			{
				r.Get("/", app.listProjectsHandler)
				r.Get("/{projectID}", app.showProjectHandler)
				r.Post("/", app.createProjectHandler)
				r.Patch("/{projectID}", app.updateProjectHandler)
				r.Delete("/{projectID}", app.deleteProjectHandler)
			}
		})

		r.Route("/products", func(r chi.Router) {
			r.Use(app.authenticate)
			{
				r.Get("/", app.listProductsHandler)
				r.Get("/{productID}", app.showProductHandler)
				r.Post("/", app.createProductHandler)
				r.Patch("/{productID}", app.updateProductHandler)
				r.Delete("/{productID}", app.deleteProductHandler)
			}
		})

		r.Route("/units", func(r chi.Router) {
			r.Use(app.authenticate)
			{
				r.Get("/", app.listUnitsHandler)
				r.Get("/{unitID}", app.showUnitHandler)
				r.Post("/", app.createUnitHandler)
				r.Patch("/{unitID}", app.updateUnitHandler)
				r.Delete("/{unitID}", app.deleteUnitHandler)
			}
		})

		r.Route("/vat_rates", func(r chi.Router) {
			r.Use(app.authenticate)
			{
				r.Get("/", app.listVatRatesHandler)
				r.Get("/{vatRateID}", app.showVatRateHandler)
				r.Post("/", app.createVatRateHandler)
				r.Patch("/{vatRateID}", app.updateVatRateHandler)
				r.Delete("/{vatRateID}", app.deleteVatRateHandler)
			}
		})

		r.Route("/invoices", func(r chi.Router) {
			r.Use(app.authenticate)
			{
				r.Get("/", app.listInvoicesHandler)
				r.Get("/{invoiceID}", app.showInvoiceHandler)
				r.Post("/", app.createInvoiceHandler)
				r.Patch("/{invoiceID}", app.updateInvoiceHandler)
				r.Delete("/{invoiceID}", app.deleteInvoiceHandler)

				r.Get("/{invoiceID}/invoice_items", app.listInvoiceItemsHandler)
				r.Get("/{invoiceID}/invoice_items/{ID}", app.showInvoiceItemHandler)
				r.Post("/{invoiceID}/invoice_items", app.createInvoiceItemHandler)
				r.Patch("/{invoiceID}/invoice_items/{ID}", app.updateInvoiceItemHandler)
				r.Delete("/{invoiceID}/invoice_items/{ID}", app.deleteInvoiceItemHandler)
			}
		})

	})

	// Return the router instance.
	return r
}
