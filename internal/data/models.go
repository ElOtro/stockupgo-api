package data

import (
	"errors"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Define a custom ErrRecordNotFound error. We'll return this from our Get() method when
// looking up a movie that doesn't exist in our database.
var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

// Create a Models struct which wraps all models.
type Models struct {
	Users         UserModel
	Organisations OrganisationModel
	BankAccounts  BankAccountModel
	Companies     CompanyModel
	Contacts      ContactModel
	Agreements    AgreementModel
	Projects      ProjectModel
	Products      ProductModel
	Units         UnitModel
	VatRates      VatRateModel
	Invoices      InvoiceModel
	InvoiceItems  InvoiceItemModel
	Helper        Helper
}

// For ease of use, we also add a New() method which returns a Models struct containing
// the initialized InvoiceModel.
func NewModels(db *pgxpool.Pool) Models {
	return Models{
		Users:         UserModel{DB: db},
		Organisations: OrganisationModel{DB: db},
		BankAccounts:  BankAccountModel{DB: db},
		Companies:     CompanyModel{DB: db},
		Contacts:      ContactModel{DB: db},
		Agreements:    AgreementModel{DB: db},
		Projects:      ProjectModel{DB: db},
		Products:      ProductModel{DB: db},
		Units:         UnitModel{DB: db},
		VatRates:      VatRateModel{DB: db},
		Invoices:      InvoiceModel{DB: db},
		InvoiceItems:  InvoiceItemModel{DB: db},
		Helper:        Helper{DB: db},
	}
}
