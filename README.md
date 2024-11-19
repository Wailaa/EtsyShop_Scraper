
# EtsyScraper


The project goal is to develop a scraper designed to gather data from Etsy shops and analyze it comprehensively. This tool enables users to monitor shop performance, gain insights into listings, demand trends, and user interest. Ultimately, it provides Etsy sellers with an invaluable online research tool to assess similar shops' performance and discover inspiration from fresh listings and products.
This porgect is build for learning purpases , to build a stateless Golang Backend 



## Environment Variables

To run this project, you will need to add the following environment variables to your project.env file, add this file in the same directory as main.go.

`POSTGRES_HOST`=host.docker.internal

`POSTGRES_USERNAME`=postgres

`POSTGRES_PASSWORD`=

`POSTGRES_DB`=

`POSTGRES_PORT`=5432

`PORT`= 8080

`CLIENT_ORIGIN`=http://localhost

`JWT_SECRET`=-

`ACCESS_TOKEN_DUARATION` 

`REFRESH_TOKEN_DUARATION`


`PROJECT_EMAIL_ADDRESS`

`MAILTRAP_SMTP_HOST`

`MAILTRAP_SMTP_USER`

`MAILTRAP_SMTP_PASS`

`MAILTRAP_SMTP_PORT`

`REDISURL`=host.docker.internal:6379

`SCRAP_SHOP_URL`=https://www.etsy.com/de-en/shop/

`SCRAP_MAX_PAGE_LIMIT`=

`PROXY_HOST_URL1`=

`PROXY_HOST_URL2`=

`PROXY_HOST_URL3`=



## Deployment

This project is built with Golang.

1. install if you dont have in  [go](https://go.dev/).


2. Ensure docker-compose is installed on your build system. For details on how to do this please click [here](https://docs.docker.com/compose/install/)

3. This application uses Mailtrap.io mailing services.You can open an account for free [here](https://mailtrap.io/).Please add the credits in project.env file to have fully working flow.
4. you can add proxy server in .env file, if not please keep it empty

5. To build a Docker image, please make sure you have obtained a copy of the repository and a working installation of Docker. Please refer to the Docker [website](https://docs.docker.com/), to learn more about how to download and install Docker.

After you cloned the repository , run the following command to build the images and run the docker container:

```bash
    docker compose up -d
```
to stop and remove container , please run:
```bash
    docker compose down
```


## Documentation

full endpoint list is available [here](https://github.com/Wailaa/EtsyShop_Scraper/blob/master/documentaions.md)

Code diagram (overview of the system's code) available [here](https://github.com/Wailaa/EtsyShop_Scraper/blob/master/diagrams/diagrams.md)

C$ diagrams available [here](https://github.com/Wailaa/EtsyShop_Scraper/blob/master/diagrams/C4_diagrams.md)

