# BankBackend
A bank application's backend

# Try it out
## Set up
Run these commands in the root directory
1. `make postgres` - to run a postgres image on docker
2. `make createdb` - to create the database

~~3. `make migrate` - to create the db schema, initial seed data will be provided on running the server~~
   
3. `make server` - to run the server

## API v1

(documentation is also available at `localhost:8080/swagger`)

The following APIs are available to the user, you may run the commands below (or use postman). 
### Create User
```sh
curl --request POST 'http://localhost:8080/v1/create_user' \
    --header 'Content-Type: application/json' \
    --data-raw '{
        "username": "<username>",
        "full_name": "<full_name>",
        "email": "<email>",
        "password": "<password>"
    }'
```
I've provided 2 users with multiple accounts: (username: peepo, password: password123) and (username: gondola, password: password123)

### Login
```sh
curl --request POST 'http://localhost:8080/v1/login_user' \
    --header 'Content-Type: application/json' \
    --data '{
        "username": "<username>",
        "password": "<password>"
    }'
```

### Create Account
Create accounts for logged in user
```sh
curl --request POST 'http://localhost:8080/v1/create_account' \
    --header 'Content-Type: application/json' \
    --header 'Authorization: Bearer <your_token>' \
    --data '{
        "owner":"<full_name>",
        "currency":"<currency>"
    }'
```

### Get Account
Get account by id
```sh
curl --request GET 'http://localhost:8080/v1/get_account/2' \
    --header 'Authorization: Bearer <your_token>'
```

### List Accounts
List accounts associated with user
```shell 
curl --request GET 'http://localhost:8080/v1/list_accounts?page_id=<start_page>&page_size=<page_size>' \
    --header 'Authorization: Bearer <your_token>'
```

### Transfer
Transfer money from logged in user's account, to any other account with matching currency
```sh
curl --request POST 'http://localhost:8080/v1/transfer_funds' \
    --header 'Content-Type: application/json' \
    --header 'Authorization: Bearer <your_token>' \
    --data '{
        "from_account_id": <from_account_id>,
        "to_account_id": <to_account_id>,
        "amount": <amount>,
        "currency": <currency>
    }'
```

### Refresh Access Token 
```shell
curl -request POST 'http://localhost:8080/v1/renew_access_token' \
    --header 'Content-Type: application/json' \
    --data '{
        "refresh_token": "<your refresh token"
    }'
```
refresh_token is acquired from the login response

## v0
### Logout
```shell
curl --header "Content-Type: application/json" \
    --request POST \
    --data '{"session_id":"<session id>"}' \
    http://localhost:8080/tokens/users/logout
```
The `session_id` can be acquired from the Login response. Note that as of yet, all logout does is invalidates the refresh token so the user can no longer renew their access token. It does not prevent access to other resources while the access token yet to expire.



Utmost thanks to Quang Pham, absolute legend. 

