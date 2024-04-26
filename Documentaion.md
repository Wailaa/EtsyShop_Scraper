
## API Reference



### Login

Used to collect a Token for a registered User.

**URL**: `/auth/login`

**Method** : `POST`

**Auth required** : NO

**Data constraints**

```json
{
  "email": "test@test.com",
  "password": 1234ssaa
}
```

**Data example**

```json
{
    "username": "Biggie@isthebest.com",
    "password": "abcd1234"
}
```

### Success Response

**Code** : `200 OK`

**Content example**

```json
{
 "token_type": "Bearer",
    "access_token": "SomeJwtToekn",
    "refresh_token": "SomeJwtRefreshToken",
    "User": {
        "Name": "Example",
        "Email": "Example@example.com",
        "Shops": [
            {
                "shop_name": "Example",
                "shop_description": "Description",
                "location": "Mars",
                "shop_total_sales": 81,
                "joined_since": "1800",
                "last_update_time": "2024-4-24",
                "admirers": 19,
                "social_media_links": [
                    "links"
                ],
                "revenue": 1234,
                "avarage_item_price": 12.4,
                "shop_member": ["Owner"],
                "shop_menu": {
                    "total_items_amount": 14,
                    "items_category": 0
                },
                "shop_reviews": {
                    "shop_rate": 5,
                    "reviews_count": 15,
                    "reviews_mentions": 0
                }
            },
        ]
    }
}
```

### Error Response

**Condition** : If 'username' and 'password' combination is wrong.

**Code** : `404 NOT FOUND`

**Content** :

```json
{
    "message": "user not found",
    "status": "fail"
}
```

### Register

Used to register new user.

**URL** : `/auth/register`

**Method** : `POST`

**Auth required** : NO

**Data constraints**

```json
{
  "first_name":"Exampe",
  "last_name":"Example",
  "email":"test@test.com",
  "password":"1234qwer",
  "password_confirm":"1234qwer",
  "subscription_type":"free"
  
}
```

### Success Response

**Code** : `200 OK`

**Content example**

```json
{
    "message": "thank you for registering, please check your email inbox",
    "status": "success"
}
```

### Error Response

**Condition** : If 'email' and is already registered.

**Code** : `409 CONFLICT`

**Content** :

```json
{
    "message": "this email is already in use",
    "status": "registraition rejected"
}
```

**Condition** : If 'password' and 'password_confirm' are not matched.

**Code** : `400 NOT FOUND`

**Content** :

```json
{
    "message": "Your password and confirmation password do not match",
    "status": "fail"
}
```
### VerifyUser

Used to verify the email after user registers.

**URL** : `/auth/verifyaccount`

**Method** : `GET`

| Parameter | Type     | Description                |
| :-------- | :------- | :------------------------- |
| `TransID` | `string` | **Required**.              |
  


**Auth required** : NO


### Success Response

**Code** : `200 OK`

**Content example**

```json
 {
     "status": "success",
     "message":"Email has been verified",
 }
```

### Error Response

**Condition** : If 'TransID' is invalid.

**Code** : `403 NOT FOUND`

**Content** :

```json
{
    "status": "fail",
     "message":"Invalid verification code or account does not exists",
}
```

**Condition** : If 'TransID' was already consumed.

**Code** : `403 NOT FOUND`

**Content** :

```json
{
    "status": "fail",
     "message":"this link is not valid anymore",
}
```

**Condition** : If 'TransID' was already consumed.

**Code** : `403 NOT FOUND`

**Content** :

```json
{
    "status": "fail",
     "message":"this link is not valid anymore",
}
```

### LogOut

Used to logout user , delete cookies, and black list JWTs.

**URL**: `/auth/logout`

**Method** : `GET`

**Auth required** : YES


### Success Response

**Code** : `200 OK`

**Content example**

```json
{
    "message": "user logged out successfully",
    "status": "success"
}
```

### Change Password

when user want to optain new password.

**URL** : `/auth/changepassword`

**Method** : `POST`

**Auth required** : YES

**Data constraints**

```json
{
    "current_password":"11111111",
    "new_password":"1111qqqq",
    "confirm_password":"1111qqqq"
}
```

### Success Response

**Code** : `200 OK`

**Content example**

```json
{
    "message": "password changed",
    "status": "registraition rejected"
}{
    "message": "user logged out successfully",
    "status": "success"
}
```

### Error Response

**Condition** : If no json was added to request body.

**Code** : `404 NOT FOUND`

**Content** :

```json
{
    "status": "fail",
     "message":"failed to fetch change password request",
}
```

**Condition** : If current_password not matched with database record.

**Code** : `404 NOT FOUND`

**Content** :

```json
{
    "status": "fail",
     "message":"password is incorrect",
}
```



### Forgot Password

when user forgot passwrod and want to get a reset request

**URL** : `/auth/forgotpassword`

**Method** : `POST`

**Auth required** : NO

**Data constraints**

```json
{
    "email_account": "test@test.com"
}
```

### Success Response

**Code** : `200 OK`

**Content example**

```json
{
    "status": "success"
}
```

### Error Response


**Condition** : If no json was added to request body.

**Code** : `404 NOT FOUND`

**Content** :

```json
{
    "status": "fail",
     "message":"failed to fetch change password request",
}
```



### Reset Password

when user apply password change.

**URL** : `/auth/resetpassword`

**Method** : `POST`

**Auth required** : NO

**Data constraints**

```json
{
    "rcp":"1111111999",
    "new_password":"1111qqqq",
    "confirm_password":"1111qqqq"
}
```

### Success Response

**Code** : `200 OK`

**Content example**

```json
{
    "status": "Password changed successfully"
}
```

### Error Response


**Condition** : If no json was added to request body.

**Code** : `404 NOT FOUND`

**Content** :

```json
{
    "status": "fail",
     "message":"failed to fetch change password request",
}
```


**Condition** : if passwords are not matched.

**Code** : `403 FORBIDDEN`

**Content** :

```json
{
    "status": "fail",
     "message": "Passwords are not the same",
}
```


**Condition**: if rcp not found or already consumed.

**Code** : `403 FORBIDDEN`

**Content** :

```json
{
    "status": "fail",
     "message": "Invalid verification code or account does not exist",
}
```




### Create Shop 

when requesting to create a shop.

**URL** : `/shop/create_shop`

**Method** : `GET`

**Auth required** : YES

**Data constraints**

```json
{
    "new_shop_name":"ShopExample"
}
```

### Success Response

**Code** : `200 OK`

**Content example**

```json
{
    "status": "success",
    "result": "shop request received successfully"
}
```

### Error Response


**Condition** : If no json was added to request body.

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "status": "fail",
    "message": "failed to get the Shop's name",
}
```


**Condition** : if shop already exists.

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "status": "fail",
    "message": "Shop already exists",
}
```



### Follow Shop 

when the user request to follow a shop.

**URL** : `/shop/follow_shop`

**Method** : `GET`

**Auth required** : YES

**Data constraints**

```json
{
    "follow_shop":"ShopName"
}
```

### Success Response

**Code** : `200 OK`

**Content example**

```json
{
    "status": "success",
    "result": "following shop"
}
```

### Error Response


**Condition** : If no JSON was added to the request body.

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "message": "EOF",
    "message": "failed to get the Shop's name",
}
```


**Condition** : if shop does not exists.

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "status": "fail",
    "message": "shop not found",
}
```

### Unfollow Shop 

when the user applies password change.

**URL** : `/shop/unfollow_shop`

**Method** : `GET`

**Auth required** : YES

**Data constraints**

```json
{
    "unfollow_shop":"ShopExample"
}
```

### Success Response

**Code** : `200 OK`

**Content example**

```json
{
    "status": "success",
    "result": "Unfollowed shop"
}
```

### Error Response


**Condition** : If no json was added to request body.

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "message": "EOF",
    "status": "fail",
}
```


**Condition** : if user request to unfollow a shop that does not exists.

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "status": "fail",
    "message": "shop not found",
}
```

### Shop

request a shop by Id.

**URL** : `/shop/{id}`

**Method** : `GET`

**Auth required** : YES


| Parameter | Type     | Description                       |
| :-------- | :------- | :-------------------------------- |
| `id`      | `string` | **Required**. Id of shop to fetch |


### Success Response

**Code** : `200 OK`

**Content example**

```json
{
    "shop_name": "Example",
    "shop_description": "example",
    "location": "mars",
    "shop_total_sales": 956,
    "joined_since": "2016",
    "last_update_time": "May 11, 2022",
    "admirers": 430,
    "social_media_links": [
        "Link_One",
        "Link_Two",
        "Linnk_Three",
    ],
    "revenue": 501827.73,
    "avarage_item_price": 90.25,
    "shop_member": [
        {
            "name": "John",
            "role": "Owner, Designer"
        },
        {
            "name": "Chtistina",
            "role": "Owner, Marketer, Photographer"
        }
    ],
    "shop_menu": {
        "total_items_amount": 52,
        "items_category": [
            {
                "category_name": "All",
                "link": "linktoCategory",
                "item_amount": 52
            },               
           
        ]
    },
    "shop_reviews": {
        "shop_rate": 4.6977,
        "reviews_count": 184,
        "reviews_mentions": [
            {
                "keyword": "quality",
                "keyword_count": 29
            },
            {
                "keyword": "shipping",
                "keyword_count": 15
            },
            {
                "keyword": "customer_service",
                "keyword_count": 26
            }
        ]
    }
}
```

### Error Response


**Condition** : if failed to get shop id.

**Code** : `400 BAD REQUEST`

**Content** :

```json

    "status": "fail",
    "message": "failed to get Shop id"
}
```


**Condition**: if shop does not  exists.

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "status": "fail",
    "message": "record not found",
}
```
## Retrieve All Items for a Shop

Retrieve a active listing  associated with a specific shop.

- **URL**: `/shop/{id}/all_items`
- **Method**: `GET`
- **Authentication required**: Yes

### Parameters

| Name     | Type     | Description                   |
|----------|----------|-------------------------------|
| `id`     | `string` | **Required**. ID of the shop |

### Response

- **Status Code**: `200 OK`
- **Content Type**: `application/json`

#### Success Response

```json
[
    {
        "Name": "Custom Order",
        "OriginalPrice": 40,
        "CurrencySymbol": "$",
        "SalePrice": 20,
        "DiscoutPercent": "50%",
        "Available": false,
        "ItemLink": "www.",
        "ListingID": 123141242,
        "PriceHistory": null
    },
    ...
]
```

### Error Response


**Condition** : if failed to get shop id.

**Code** : `400 BAD REQUEST`

**Content** :

```json

    "status": "fail",
    "message": "failed to get Shop id"
}
```


**Condition**: if shop does not  exists.

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "status": "fail",
    "message": "record not found",
}
```

## Retrieve All Sold items for a Shop

Retrieve all sold items  associated with a specific shop.

- **URL**: `/shop/{id}/all_sold_items`
- **Method**: `GET`
- **Authentication required**: Yes

### Parameters

| Name     | Type     | Description                   |
|----------|----------|-------------------------------|
| `id`     | `string` | **Required**. ID of the shop |

### Response

- **Status Code**: `200 OK`
- **Content Type**: `application/json`

#### Success Response

```json
[
    {
        "Name": "custom item Pipe Lamp",
        "ItemID": 1,
        "OriginalPrice": 28.8,
        "CurrencySymbol": "€",
        "SalePrice": -1,
        "DiscoutPercent": "",
        "ItemLink": "www.ExampleLink.com",
        "Available": true,
        "SoldQauntity": 60
    },
    ...
]
```

### Error Response


**Condition** : if failed to get shop id.

**Code** : `400 BAD REQUEST`

**Content** :

```json

    "status": "fail",
    "message": "failed to get Shop id"
}
```


**Condition**: if shop does not  exists.

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "status": "fail",
    "message": "record not found",
}
```

## Get last 30 days statistics 

Generate last 30 days selling history for a Shop.


- **URL**: `shop/stats/{id}/lastThirtyDays`
- **Method**: `GET`
- **Authentication required**: Yes

### Parameters

| Name     | Type     | Description                   |
|----------|----------|-------------------------------|
| `id`     | `string` | **Required**. ID of the shop |

### Response

- **Status Code**: `200 OK`
- **Content Type**: `application/json`

#### Success Response

```json
"stats": {
        "2024-04-20": {
            "total_sales": 453,
            "Items": [
                {
                    "Name": "item1",
                    "OriginalPrice": 102.96,
                    "CurrencySymbol": "€",
                    "SalePrice": -1,
                    "DiscoutPercent": "",
                    "Available": true,
                    "ItemLink": "examplelink.com",
                    "ListingID": 1159304220,
                    "PriceHistory": null
                }
            ]
        },
    ...
]
```

### Error Response


**Condition** : if failed to get shop id.

**Code** : `400 BAD REQUEST`

**Content** :

```json

    "status": "fail",
    "message": "failed to get Shop id"
}
```


**Condition**: if shop does not  exists.

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "status": "fail",
    "message": "record not found",
}
```

## Get last 90 days statistics 

Generate last 90 days selling history for a Shop.


- **URL**: `shop/stats/{id}/lastThreeMonths`
- **Method**: `GET`
- **Authentication required**: Yes

### Parameters

| Name     | Type     | Description                   |
|----------|----------|-------------------------------|
| `id`     | `string` | **Required**. ID of the shop |

### Response

- **Status Code**: `200 OK`
- **Content Type**: `application/json`

#### Success Response

```json
"stats": {
        "2024-04-20": {
            "total_sales": 453,
            "Items": [
                {
                    "Name": "item1",
                    "OriginalPrice": 102.96,
                    "CurrencySymbol": "€",
                    "SalePrice": -1,
                    "DiscoutPercent": "",
                    "Available": true,
                    "ItemLink": "examplelink.com",
                    "ListingID": 1159304220,
                    "PriceHistory": null
                }
            ]
        },
    ...
]
```

### Error Response


**Condition** : if failed to get shop id.

**Code** : `400 BAD REQUEST`

**Content** :

```json

    "status": "fail",
    "message": "failed to get Shop id"
}
```


**Condition**: if shop does not  exists.

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "status": "fail",
    "message": "record not found",
}
```

## items Count

Get available and sold out items by shop.


- **URL**: `/shop/{id}/items_count`
- **Method**: `GET`
- **Authentication required**: Yes

### Parameters

| Name     | Type     | Description                   |
|----------|----------|-------------------------------|
| `id`     | `string` | **Required**. ID of the shop |

### Response

- **Status Code**: `200 OK`
- **Content Type**: `application/json`

#### Success Response

```json
{
    "Available": 20,
    "OutOfProduction": 50
}
```

### Error Response


**Condition** : if failed to get shop id.

**Code** : `400 BAD REQUEST`

**Content** :

```json

    "status": "fail",
    "message": "failed to get Shop id"
}
```


**Condition**: if shop does not  exists.

**Code** : `400 BAD REQUEST`

**Content** :

```json
{
    "status": "fail",
    "message": "record not found",
}
```


