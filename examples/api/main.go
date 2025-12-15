package main

// Swagger Petstore - OpenAPI 3.0
//
// This is a sample Pet Store Server based on the OpenAPI 3.0 specification.
// You can find out more about Swagger at https://swagger.io.
//
// !api 3.0.3
// !info "Swagger Petstore - OpenAPI 3.0" v1.0.27 "This is a sample Pet Store Server based on the OpenAPI 3.0 specification. You can find out more about Swagger at [https://swagger.io](https://swagger.io). In the third iteration of the pet store, we've switched to the design first approach! You can now help us improve the API whether it's by making changes to the definition itself or to the code. That way, with time, we can improve the API in general, and expose some of the new features in OAS3."
// !contact "" <apiteam@swagger.io>
// !license Apache-2.0 https://www.apache.org/licenses/LICENSE-2.0.html
// !tos https://swagger.io/terms/
// !externalDocs https://swagger.io "Find out more about Swagger"
// !link "The Pet Store repository" https://github.com/swagger-api/swagger-petstore
// !link "The source API definition for the Pet Store" https://github.com/swagger-api/swagger-petstore/blob/master/src/main/resources/openapi.yaml
// !security petstore_auth:oauth2 "OAuth2 authentication" https://petstore3.swagger.io/oauth/authorize
// !scope petstore_auth write:pets "modify pets in your account"
// !scope petstore_auth read:pets "read your pets"
// !security api_key:apiKey:header "API Key authentication"
// !server https://petstore3.swagger.io/api/v3 "Petstore server"
// !server http://localhost:8080/api/v3 "Local development"
// !tag pet "Everything about your Pets"
// !tag store "Access to Petstore orders"
// !tag user "Operations about user"
func main() {}

// ============================================================================
// PET OPERATIONS
// ============================================================================

// UpdatePet updates an existing pet by ID.
//
// !PUT /pet -> updatePet "Update an existing pet" #pet
// !secure petstore_auth api_key
// !body Pet "Update an existent pet in the store" required
// !ok Pet "Successful operation"
// !error 400 ApiResponse "Invalid ID supplied"
// !error 404 ApiResponse "Pet not found"
// !error 405 ApiResponse "Validation exception"
func UpdatePet() {}

// AddPet adds a new pet to the store.
//
// !POST /pet -> addPet "Add a new pet to the store" #pet
// !secure petstore_auth api_key
// !body Pet "Create a new pet in the store" required
// !ok 200 Pet "Successful operation"
// !error 405 ApiResponse "Invalid input"
func AddPet() {}

// FindPetsByStatus finds pets by status.
// Multiple status values can be provided with comma separated strings.
//
// !GET /pet/findByStatus -> findPetsByStatus "Finds Pets by status" #pet
// !secure petstore_auth api_key
// !query status:string "Status values that need to be considered for filter" default=available
// !ok Pet[] "Successful operation"
// !error 400 ApiResponse "Invalid status value"
func FindPetsByStatus() {}

// FindPetsByTags finds pets by tags.
// Multiple tags can be provided with comma separated strings.
//
// !GET /pet/findByTags -> findPetsByTags "Finds Pets by tags" #pet
// !secure petstore_auth api_key
// !query tags:string "Tags to filter by"
// !ok Pet[] "Successful operation"
// !error 400 ApiResponse "Invalid tag value"
func FindPetsByTags() {}

// GetPetById finds pet by ID.
// Returns a single pet.
//
// !GET /pet/{petId} -> getPetById "Find pet by ID" #pet
// !secure api_key petstore_auth
// !path petId:int64 "ID of pet to return" required
// !ok Pet "Successful operation"
// !error 400 ApiResponse "Invalid ID supplied"
// !error 404 ApiResponse "Pet not found"
func GetPetById() {}

// UpdatePetWithForm updates a pet in the store with form data.
//
// !POST /pet/{petId} -> updatePetWithForm "Updates a pet in the store with form data" #pet
// !secure petstore_auth api_key
// !path petId:int64 "ID of pet that needs to be updated" required
// !query name:string "Name of pet that needs to be updated"
// !query status:string "Status of pet that needs to be updated"
// !ok ApiResponse "Successful operation"
// !error 405 ApiResponse "Invalid input"
func UpdatePetWithForm() {}

// DeletePet deletes a pet.
//
// !DELETE /pet/{petId} -> deletePet "Deletes a pet" #pet
// !secure petstore_auth api_key
// !header api_key:string "API key for authentication"
// !path petId:int64 "Pet id to delete" required
// !ok ApiResponse "Successful operation"
// !error 400 ApiResponse "Invalid pet value"
func DeletePet() {}

// UploadFile uploads an image for a pet.
//
// !POST /pet/{petId}/uploadImage -> uploadFile "Uploads an image" #pet
// !secure petstore_auth api_key
// !path petId:int64 "ID of pet to update" required
// !query additionalMetadata:string "Additional Metadata"
// !body FileUploadRequest "Image file to upload"
// !ok ApiResponse "Successful operation"
func UploadFile() {}

// ============================================================================
// STORE OPERATIONS
// ============================================================================

// GetInventory returns pet inventories by status.
// Returns a map of status codes to quantities.
//
// !GET /store/inventory -> getInventory "Returns pet inventories by status" #store
// !secure api_key
// !ok InventoryResponse "Successful operation"
func GetInventory() {}

// PlaceOrder places an order for a pet.
//
// !POST /store/order -> placeOrder "Place an order for a pet" #store
// !body Order "Place order for a pet" required
// !ok 200 Order "Successful operation"
// !error 405 ApiResponse "Invalid input"
func PlaceOrder() {}

// GetOrderById finds purchase order by ID.
// For valid response try integer IDs with value <= 5 or > 10. Other values will generate exceptions.
//
// !GET /store/order/{orderId} -> getOrderById "Find purchase order by ID" #store
// !path orderId:int64 "ID of order that needs to be fetched" required
// !ok Order "Successful operation"
// !error 400 ApiResponse "Invalid ID supplied"
// !error 404 ApiResponse "Order not found"
func GetOrderById() {}

// DeleteOrder deletes purchase order by ID.
// For valid response try integer IDs with value < 1000. Anything above 1000 or nonintegers will generate API errors.
//
// !DELETE /store/order/{orderId} -> deleteOrder "Delete purchase order by ID" #store
// !path orderId:int64 "ID of the order that needs to be deleted" required
// !ok ApiResponse "Successful operation"
// !error 400 ApiResponse "Invalid ID supplied"
// !error 404 ApiResponse "Order not found"
func DeleteOrder() {}

// ============================================================================
// USER OPERATIONS
// ============================================================================

// CreateUser creates a new user.
// This can only be done by the logged in user.
//
// !POST /user -> createUser "Create user" #user
// !body User "Created user object" required
// !ok User "Successful operation"
func CreateUser() {}

// CreateUsersWithListInput creates a list of users with given input array.
//
// !POST /user/createWithList -> createUsersWithListInput "Creates list of users with given input array" #user
// !body User[] "List of user objects" required
// !ok User "Successful operation"
func CreateUsersWithListInput() {}

// LoginUser logs user into the system.
//
// !GET /user/login -> loginUser "Logs user into the system" #user
// !query username:string "The user name for login"
// !query password:string "The password for login in clear text"
// !ok LoginResponse "Successful operation"
// !error 400 ApiResponse "Invalid username/password supplied"
func LoginUser() {}

// LogoutUser logs out current logged in user session.
//
// !GET /user/logout -> logoutUser "Logs out current logged in user session" #user
// !ok ApiResponse "Successful operation"
func LogoutUser() {}

// GetUserByName gets user by user name.
//
// !GET /user/{username} -> getUserByName "Get user by user name" #user
// !path username:string "The name that needs to be fetched. Use user1 for testing." required
// !ok User "Successful operation"
// !error 400 ApiResponse "Invalid username supplied"
// !error 404 ApiResponse "User not found"
func GetUserByName() {}

// UpdateUser updates user.
// This can only be done by the logged in user.
//
// !PUT /user/{username} -> updateUser "Update user" #user
// !path username:string "Name that needs to be updated" required
// !body User "Update an existent user in the store" required
// !ok User "Successful operation"
func UpdateUser() {}

// DeleteUser deletes user.
// This can only be done by the logged in user.
//
// !DELETE /user/{username} -> deleteUser "Delete user" #user
// !path username:string "The name that needs to be deleted" required
// !ok ApiResponse "Successful operation"
// !error 400 ApiResponse "Invalid username supplied"
// !error 404 ApiResponse "User not found"
func DeleteUser() {}

// ============================================================================
// MODELS
// ============================================================================

// Pet represents a pet in the store.
// !model "A pet for sale in the pet store"
type Pet struct {
	// !field id:int64 "Unique identifier for the pet" example=10
	ID int64 `json:"id,omitempty"`

	// !field name:string "Name of the pet" required example="doggie"
	Name string `json:"name"`

	// !field category:Category "Category of the pet"
	Category *Category `json:"category,omitempty"`

	// !field photoUrls:string[] "List of photo URLs" required
	PhotoUrls []string `json:"photoUrls"`

	// !field tags:Tag[] "Tags associated with the pet"
	Tags []Tag `json:"tags,omitempty"`

	// !field status:string "Pet status in the store" example="available"
	Status string `json:"status,omitempty"`
}

// Category represents a pet category.
// !model "A category for a pet"
type Category struct {
	// !field id:int64 "Unique identifier" example=1
	ID int64 `json:"id,omitempty"`

	// !field name:string "Category name" example="Dogs"
	Name string `json:"name,omitempty"`
}

// Tag represents a tag for pets.
// !model "A tag for a pet"
type Tag struct {
	// !field id:int64 "Unique identifier" example=0
	ID int64 `json:"id,omitempty"`

	// !field name:string "Tag name" example="string"
	Name string `json:"name,omitempty"`
}

// Order represents a purchase order.
// !model "An order for a pet"
type Order struct {
	// !field id:int64 "Unique identifier" example=10
	ID int64 `json:"id,omitempty"`

	// !field petId:int64 "ID of the pet being ordered" example=198772
	PetID int64 `json:"petId,omitempty"`

	// !field quantity:integer "Number of pets ordered" example=7
	Quantity int `json:"quantity,omitempty"`

	// !field shipDate:string "Estimated ship date" example="2024-01-15T10:30:00Z"
	ShipDate string `json:"shipDate,omitempty"`

	// !field status:string "Order status" example="approved"
	Status string `json:"status,omitempty"`

	// !field complete:boolean "Whether the order is complete" example=true
	Complete bool `json:"complete,omitempty"`
}

// User represents a user in the system.
// !model "A User who can purchase pets"
type User struct {
	// !field id:int64 "Unique identifier" example=10
	ID int64 `json:"id,omitempty"`

	// !field username:string "Username for login" example="theUser"
	Username string `json:"username,omitempty"`

	// !field firstName:string "First name" example="John"
	FirstName string `json:"firstName,omitempty"`

	// !field lastName:string "Last name" example="James"
	LastName string `json:"lastName,omitempty"`

	// !field email:string "Email address" example="john@email.com"
	Email string `json:"email,omitempty"`

	// !field password:string "Password" example="12345"
	Password string `json:"password,omitempty"`

	// !field phone:string "Phone number" example="12345"
	Phone string `json:"phone,omitempty"`

	// !field userStatus:integer "User status" example=1
	UserStatus int `json:"userStatus,omitempty"`
}

// ApiResponse represents a standard API response.
// !model "API response object"
type ApiResponse struct {
	// !field code:integer "Response code" example=200
	Code int `json:"code,omitempty"`

	// !field type:string "Response type" example="success"
	Type string `json:"type,omitempty"`

	// !field message:string "Response message" example="Operation successful"
	Message string `json:"message,omitempty"`
}

// Address represents a shipping address.
// !model "Shipping address"
type Address struct {
	// !field street:string "Street address" example="437 Lytton"
	Street string `json:"street,omitempty"`

	// !field city:string "City" example="Palo Alto"
	City string `json:"city,omitempty"`

	// !field state:string "State" example="CA"
	State string `json:"state,omitempty"`

	// !field zip:string "ZIP code" example="94301"
	Zip string `json:"zip,omitempty"`
}

// Customer represents a customer.
// !model "A customer who purchases pets"
type Customer struct {
	// !field id:int64 "Unique identifier" example=100000
	ID int64 `json:"id,omitempty"`

	// !field username:string "Username" example="fehguy"
	Username string `json:"username,omitempty"`

	// !field address:Address[] "Customer addresses"
	Address []Address `json:"address,omitempty"`
}

// InventoryResponse represents inventory counts by status.
// !model "Inventory counts by pet status"
type InventoryResponse struct {
	// !field available:integer "Number of available pets" example=50
	Available int `json:"available,omitempty"`

	// !field pending:integer "Number of pending pets" example=10
	Pending int `json:"pending,omitempty"`

	// !field sold:integer "Number of sold pets" example=100
	Sold int `json:"sold,omitempty"`
}

// LoginResponse represents a successful login response.
// !model "Login response with session token"
type LoginResponse struct {
	// !field token:string "Session token" example="logged in user session:1234567890"
	Token string `json:"token,omitempty"`

	// !field expiresAfter:string "Session expiration time" example="2024-01-15T11:30:00Z"
	ExpiresAfter string `json:"expiresAfter,omitempty"`

	// !field rate-limit:integer "Calls per hour allowed by the user" example=5000
	RateLimit int `json:"rate-limit,omitempty"`
}

// FileUploadRequest represents a file upload request.
// !model "File upload request body"
type FileUploadRequest struct {
	// !field file:string "Binary file content"
	File string `json:"file,omitempty"`
}
