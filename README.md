# BankBackend
A bank application's backend

# Try it out
## Set up
Run these commands in the root directory
1. `make postgres` - to run a postgres image on docker
2. `make createdb` - to create the database
3. `make migrate` - to create the db schema, initial seed data will be provided on running the server
4. `make server` - to run the server

## API
The following APIs are available to the user, you may run the commands below (or use postman). 
### Create User
```shell
curl --header "Content-Type: application/json" \
    --request POST \
    --data '{"username":"<username>", "full_name":"<full name>", "email":"<email>", "password":"password"}' \
    http://localhost:8080/users
```
I've provided 2 users with multiple accounts: (username: peepo, password: password123) and (username: gondola, password: password123)

### Login
```shell
curl --header "Content-Type: application/json" \
    --request POST \
    --data '{"username":"<username>", "password":"<password>"}' \
    http://localhost:8080/users/login
```

### Create Account
Create accounts for logged in user
```shell
 curl --header "Content-Type: application/json" \
    --header "Authorization: Bearer <your header token>" \
    --request POST \
    --data '{"currency":"<chosen currency>"}' \
    http://localhost:8080/accounts
```

### #Get Accounts
Get accounts associated with user
```shell 
curl --header "Content-Type: application/json" \
    --header "Authorization: Bearer <your header token>" \
    --request GET \
    http://localhost:8080/accounts/?page_id=1&page_size=10
```

### Transfer
Transfer money from logged in user's account, to any other account with matching currency
```shell
curl --header "Content-Type: application/json" \
    --header "Authorization: Bearer <your header token>" \
    --request POST \
    --data '{"from_account_id": "<from account id>", "to_account_id": "<to account id>", "amount":"<amount>", "currency": "<currency>"}' \
    http://localhost:8080/transfers 
```

### Refresh Access Token 
```shell
curl --header "Content-Type: application/json" \
    --request POST \
    --data '{"refresh_token":"<refresh token>"}' \
    http://localhost:8080/tokens/renew_access
```

### Logout
```shell
curl --header "Content-Type: application/json" \
    --request POST \
    --data '{"session_id":"<session id>"}' \
    http://localhost:8080/tokens/users/logout
```
The `session_id` can be acquired from the Login response. Note that as of yet, all logout does is invalidates the refresh token so the user can no longer renew their access token. It does not prevent access to other resources while the access token yet to expire.



Utmost thanks to Quang Pham, absolute legend. 

