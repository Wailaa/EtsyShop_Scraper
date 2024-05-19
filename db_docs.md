
# Introduction:

In this self-assessment essay, I will reflect on my experience in developing a project that involved building a Postgres database using the Golang programming language, Gorm as the ORM (Object-Relational Mapping) tool, and Gin for web framework implementation. 
I have a hobby were i create furniture and art figures using copper , wood and light installations. Researching the market for my furniture  crucial step in assessing the viability of turning the hobby into a profitable venture. so i was looking online to check if there is demand , similar products available, sellers , how to monetize and process data.and that is how i came with the idea to build this project.

 This project aimed to provide tool for etsy sellers to follow similar shop and track their performance and sales,be inspired by other sellers and explore new Listings ,check if there is sales to be made.
. The architecture of the application involved Golang backend served as the core component responsible for handling business logic, data processing, and interfacing with the database.The Gorm library was utilized as the ORM tool to abstract database operations and facilitate seamless communication between the Golang application and the Postgres database.

## Technical Skills

In the course of this project, I have developed a solid understanding of Golang, including its syntax, standard library, and concurrency features. My familiarity with Gorm allowed me to effectively manage database interactions, such as modeling database schema and defining relationships between tables. Additionally, my experience with Gin enabled me to build RESTful APIs and handle HTTP requests/responses efficiently. Furthermore, I demonstrated basic knowlage in SQL and Postgres, particularly in database schema design, querying, and optimization.

Project Implementation:

During the implementation phase, I followed a systematic approach to build the Postgres database schema starting from two main tables , accounts and shops,these tables formed the cornerstone of the database .As the project progressed, I further refined the schema by establishing additional relationships and breaking down the interactions between these core tables into more granular components.
The accounts table served as the central table for user information, storing essential details such as user IDs, usernames, email addresses, and authentication credentials. The shops table represented individual Etsy shops associated with each user account. It stored shop-specific data, including shop IDs, shop names, descriptions, menus and items . This table formed the backbone of the application's shop management functionality and the creation of data relations along the way.

Achievements:

Using Gorm, several advanced features were successfully integrated into the project to make sql complex quering easier, such as the use of the method Preload() as an example of eager loading of associated data, streamlining complex SQL queries involving multiple related tables. the biuld support to maintain transactional integrity during critical database operations, such as bulk updates or complex data manipulations.

Lessons Learned:

Working on this project provided valuable insights and lessons. I gained a better understanding of relational DB and ORM Utilization. 

Future Goals:

Moving forward, I aim to further enhance my skills in Postgres and  Object Relational Mapping Database, and database development. I plan to seek more like minded fellow students who have interest in developing the app further . Additionally, I intend to apply the knowledge and experience gained from this project to future projects.


Conclusion:

In conclusion, the experience of building a Postgres database using Golang, Gorm, and Gin was both challenging and rewarding. It provided an opportunity to expand my technical skills, overcome obstacles, and deliver a functional solution. the code provide in my opinion more insights of  the journey while developing the app, and the challenges encountered alomg the way.

this is the ERM model showing the relation between keys.
![Scraper_ERM_model](https://github.com/Wailaa/EtsyShop_Scraper/assets/45070102/d81d415d-9d88-47b8-9f68-59141ddad042)



![Scraper_ERM_model](https://github.com/Wailaa/EtsyShop_Scraper/assets/45070102/81b15f8c-5ea7-4274-bbed-b56f6dad4f21)
