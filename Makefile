# Generate Swagger docs
swag:
	@echo "Generating Swagger documentation..."
	swag init --parseDependency --parseInternal
