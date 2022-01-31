## App

This is an application that is used to create and store your invoices even if you have many organisations and bank accounts.

Implementation in Ruby on Rails <a href="https://github.com/ElOtro/stockuprb-api">here</a>

## Dependencies

- Pgx
- Chi
- Zerolog
- PostgreSQL
- JWT

## How to run

- create PostgreSQL database "stockup_dev"
- check settings in makefile 
- run "make migrations/up"

If you want to fill database tables with test data, run:  "go run ./cmd/api -seed"

## FAQ

Why do I use the jsonb type in bank_accounts, contacts? 

Let's think about a bank account. You maybe have multiple bank accounts. The one is a national currency account, other is a dollar account. They definitely may have different attributes, so to make your life easier, it would be nice define them in the jsonb type.

What are start_at, end_at in agreements for?

To store "history" changing company contacts.

## TODO

- Dockerize
- Project stages
- Printable invoice in PDF