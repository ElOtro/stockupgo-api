package data

import (
	"errors"
	"math/rand"
	"strconv"
	"time"

	"github.com/ElOtro/stockup-api/internal/faker"
	"github.com/ElOtro/stockup-api/internal/validator"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
)

// Define a Seed struct type which wraps a pgx.Conn connection pool.
type Seed struct {
	DB     *pgxpool.Pool
	Logger *zerolog.Logger
	Models
}

func randomInt(i int) int {
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(i)
}

// Create fake organisation.
func (s Seed) CreateOrganisations() error {

	for i := 0; i < 3; i++ {
		input := faker.NewCompany()

		organisation := Organisation{
			Name:       input.Name,
			FullName:   input.FullName,
			CEO:        input.CEO,
			CEOTitle:   "CEO",
			CFO:        input.CFO,
			CFOTitle:   "CFO",
			IsVatPayer: i%2 == 0,
			Details: &OrganisationDetails{
				INN:     input.INN,
				KPP:     input.INN,
				OGRN:    input.INN,
				Address: input.Address,
			},
		}

		// Initialize a new Validator instance.
		v := validator.New()

		// Call the validate function and return a response containing the errors if
		// any of the checks fail.
		if ValidateOrganisation(v, &organisation); !v.Valid() {
			for _, err := range v.Errors {
				s.Logger.Info().Msg(err)
			}
		}

		err := s.Organisations.Insert(&organisation)
		if err != nil {
			s.Logger.Err(err)
		}

		bankAccount := BankAccount{}
		if organisation.ID > 0 {

			bankAccount = BankAccount{
				Name: "Test",
				Details: &BankAccountDetails{
					BIK:         "1234567890",
					Account:     "1234567890",
					INN:         "1234567890",
					KPP:         "1234567890",
					CorrAccount: "1234567890",
				},
			}
			if ValidateBankAccount(v, &bankAccount); !v.Valid() {
				for _, err := range v.Errors {
					s.Logger.Info().Msg(err)
				}
			}

			err = s.BankAccounts.Insert(organisation.ID, &bankAccount)
			if err != nil {
				s.Logger.Err(err)
			}
		}

	}

	return nil

}

// Create fake vat_rates.
func (s Seed) CreateVats() error {

	vatRates := []VatRate{
		VatRate{
			IsActive:  true,
			IsDefault: false,
			Rate:      0,
			Name:      "0%",
		},
		VatRate{
			IsActive:  true,
			IsDefault: true,
			Rate:      0,
			Name:      "Без НДС",
		},
		VatRate{
			IsActive:  true,
			IsDefault: false,
			Rate:      10,
			Name:      "10%",
		},
		VatRate{
			IsActive:  true,
			IsDefault: true,
			Rate:      20,
			Name:      "20%",
		},
	}

	for _, v := range vatRates {
		err := s.VatRates.Insert(&v)
		if err != nil {
			s.Logger.Err(err)
		}
	}

	return nil
}

// Create fake units.
func (s Seed) CreateUnits() error {

	units := []Unit{
		Unit{Name: "шт."},
		Unit{Name: "час"},
	}

	for _, v := range units {
		err := s.Units.Insert(&v)
		if err != nil {
			s.Logger.Err(err)
		}
	}

	return nil
}

// Create fake company.
func (s Seed) CreateCompanies() error {

	for i := 0; i < 10; i++ {

		input := faker.NewCompany()
		company := Company{
			Name:        input.Name,
			FullName:    input.FullName,
			CompanyType: 1,
			Details: &CompanyDetails{
				INN:     input.INN,
				KPP:     input.INN,
				OGRN:    input.INN,
				Address: input.Address,
			},
		}

		// Initialize a new Validator instance.
		v := validator.New()

		// Call the validate function and return a response containing the errors if
		// any of the checks fail.
		if ValidateCompany(v, &company); !v.Valid() {
			for _, err := range v.Errors {
				s.Logger.Info().Msg(err)
			}
		}

		err := s.Companies.Insert(&company)
		if err != nil {
			s.Logger.Err(err)
		}

		if company.ID > 0 {
			err = s.CreateContacts(company.ID)
			if err != nil {
				s.Logger.Err(err)
			}
			err = s.CreateAgreements(company.ID)
			if err != nil {
				s.Logger.Err(err)
			}
		}
	}

	return nil

}

// Create fake contacts.
func (s Seed) CreateContacts(companyID int64) error {
	// Initialize a new Validator instance.
	v := validator.New()

	for i := 0; i < 2; i++ {
		input := faker.NewPerson(i%2 == 0)
		var role int
		var title string
		start := time.Now()
		if i%2 == 0 {
			role = 1
			title = "CEO"
		} else {
			role = 2
			title = "CFO"
		}

		contact := Contact{
			Role:    role,
			Title:   title,
			Name:    input.Name,
			Phone:   input.Phone,
			Email:   input.Email,
			StartAt: &start,
		}

		if ValidateContact(v, &contact); !v.Valid() {
			for _, err := range v.Errors {
				s.Logger.Info().Msg(err)
			}
		}

		err := s.Contacts.Insert(companyID, &contact)
		if err != nil {
			s.Logger.Err(err)
		}

	}

	return nil

}

// Create fake contacts.
func (s Seed) CreateAgreements(companyID int64) error {

	for i := 0; i < 5; i++ {
		input := faker.NewAgreement()
		agreement := Agreement{
			CompanyID: companyID,
			Name:      input.Name,
			StartAt:   &input.StartAt,
		}

		// Initialize a new Validator instance.
		v := validator.New()

		// Call the validate function and return a response containing the errors if
		// any of the checks fail.
		if ValidateAgreement(v, &agreement); !v.Valid() {
			for _, err := range v.Errors {
				s.Logger.Info().Msg(err)
			}
		}

		err := s.Agreements.Insert(&agreement)
		if err != nil {
			s.Logger.Err(err)
		}
	}

	return nil
}

// Create fake product.
func (s Seed) CreateProducts() error {
	fproducts := faker.ProductList()
	vatRateIDs, err := s.Helper.pluckIDs("vat_rates")
	if err != nil {
		return err
	}

	unitIDs, err := s.Helper.pluckIDs("units")
	if err != nil {
		return err
	}

	for _, p := range fproducts {
		product := Product{
			IsActive:    true,
			ProductType: 1,
			Name:        p.Name,
			Description: p.Description,
			SKU:         p.SKU,
			Price:       p.Price,
			VatRateID:   &vatRateIDs[randomInt(len(vatRateIDs))],
			UnitID:      &unitIDs[randomInt(len(unitIDs))],
		}

		// Initialize a new Validator instance.
		v := validator.New()

		// Call the validate function and return a response containing the errors if
		// any of the checks fail.
		if ValidateProduct(v, &product); !v.Valid() {
			for _, err := range v.Errors {
				s.Logger.Info().Msg(err)
			}
			return err
		}

		err := s.Products.Insert(&product)
		if err != nil {
			return err
		}
	}

	return nil
}

// Create fake invoice.
func (s Seed) CreateInvoices() error {
	organisationIDs, err := s.Helper.pluckIDs("organisations")
	if err != nil {
		return err
	}

	for _, organisationID := range organisationIDs {
		invoiceNumber := 0
		filters := CompanyFilters{Name: ""}
		pagination := Pagination{Page: 1, Limit: 1000, Sort: "id", SortSafelist: []string{"id"}}
		companies, _, err := s.Companies.GetAll(filters, pagination)
		if err != nil {
			return err
		}

		for _, v := range companies {
			agreementFilters := AgreementFilters{CompanyID: v.ID}
			pagination := Pagination{Page: 1, Limit: 1000, Sort: "id", SortSafelist: []string{"id"}}
			agreements, _, err := s.Agreements.GetAll(agreementFilters, pagination)
			if err != nil {
				return err
			}
			var agreement *Agreement
			if len(agreements) > 0 {
				agreement = agreements[randomInt(len(agreements))]
			}
			for i := 0; i < 5; i++ {
				invoiceNumber += 1
				// get bank_accounts
				bankAccounts, err := s.BankAccounts.GetAll(organisationID)
				if err != nil {
					return err
				}
				var bankAccount *BankAccount
				var bankAccountID int64
				if len(bankAccounts) > 0 {
					bankAccount = bankAccounts[0]
					bankAccountID = bankAccount.ID
				}

				invoice := Invoice{
					IsActive:       true,
					Date:           time.Now(),
					Number:         strconv.Itoa(invoiceNumber),
					OrganisationID: organisationID,
					CompanyID:      v.ID,
					AgreementID:    agreement.ID,
				}

				if bankAccountID > 0 {
					invoice.BankAccountID = bankAccountID
				}

				// Initialize a new Validator instance.
				v := validator.New()

				// Call the validate function and return a response containing the errors if
				// any of the checks fail.
				if ValidateInvoice(v, &invoice); !v.Valid() {
					for _, err := range v.Errors {
						s.Logger.Info().Msg(err)
					}
					return errors.New("invoice is not valid")
				}

				// insert to base new Invoice
				err = s.Invoices.Insert(&invoice)
				if err != nil {
					return err
				}

				if invoice.ID > 0 {
					err = s.CreateInvoiceItems(invoice.ID)
					if err != nil {
						return err
					}

					err = s.Invoices.UpdateTotals(invoice.ID)
					if err != nil {
						return err
					}
				}

			}
		}
	}

	return nil
}

// Create fake invoice.
func (s Seed) CreateInvoiceItems(invoiceID int64) error {
	products, err := s.Products.GetAll()
	if err != nil {
		return err
	}

	for i := 1; i < 4; i++ {
		var product *Product
		if len(products) > 0 {
			product = products[randomInt(len(products))]
		}
		if product != nil {
			quantity := float64(randomInt(10))
			amount := float64(quantity) * product.Price
			vat := 0.0
			if product.VatRate.Rate > 0 {
				vat = amount * (product.VatRate.Rate / 100)
			}
			invoiceItem := InvoiceItem{
				Position:    i,
				ProductID:   product.ID,
				Description: product.Description,
				UnitID:      product.Unit.ID,
				Quantity:    quantity,
				Price:       product.Price,
				Amount:      amount,
				VatRateID:   product.VatRate.ID,
				Vat:         vat,
			}

			v := validator.New()

			// Call the validate function and return a response containing the errors if
			// any of the checks fail.
			if ValidateInvoiceItem(v, &invoiceItem); !v.Valid() {
				for _, err := range v.Errors {
					s.Logger.Info().Msg(err)
				}
				return errors.New("invoiceItem is not valid")
			}

			// insert to base new Invoice
			err = s.InvoiceItems.Insert(invoiceID, &invoiceItem)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s Seed) Seed() {

	// create organisations
	err := s.CreateOrganisations()
	if err != nil {
		s.Logger.Err(err)
	}
	// create vat_rates
	err = s.CreateVats()
	if err != nil {
		s.Logger.Err(err)
	}
	// create units
	err = s.CreateUnits()
	if err != nil {
		s.Logger.Err(err)
	}
	// create companies
	err = s.CreateCompanies()
	if err != nil {
		s.Logger.Err(err)
	}
	// create products
	err = s.CreateProducts()
	if err != nil {
		s.Logger.Err(err)
	}
	// create invoices
	err = s.CreateInvoices()
	if err != nil {
		s.Logger.Err(err)
	}

}
