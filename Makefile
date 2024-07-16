gate:
	cd ./gateway && go run main.go

user:
	cd ./user_service && go run main.go

user_test:
	cd ./user_service && go test ./... -v

wallet:
	cd ./wallet_service && go run main.go

wallet_test:
	cd ./wallet_service && go test ./... -v