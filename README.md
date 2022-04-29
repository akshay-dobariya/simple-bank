# install migration tool for golang
brew install golang-migrate

# -ext extenstion, -dir target dir -seq sequential
migrate create -ext sql -dir db/migrations -seq init_schema

# installing sqlc
brew install sqlc