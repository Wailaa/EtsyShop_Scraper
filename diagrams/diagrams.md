
### Diagrams

1. the mydiagrans.puml file was generated with goplantuml package, click [here](https://github.com/jfeliu007/goplantuml?tab=readme-ov-files) for installion instructions.run the following command to generate .puml file for the code base:
```bash
    goplantuml -output mydiagrams.puml -recursive .
```

2. the UML diagram was created with PlantUML tool, click [here](https://plantuml.com/starting) for more information

the PUML diagram below describes particularly the management and interaction between different components such as shops, items, user accounts, user authentication and sales data.

## Namespaces and Classes

### collector namespace: 
    Contains the DefaultCollector class, which likely handles web scraping or data collection. The colly.Collector class appears to be used here, suggesting the use of the Colly scraping framework.

### controllers namespace: 
    Contains a variety of classes related to different functions such as shop management, user registration, sales statistics, and account operations:

    Shop class contains various operations and methods for managing shops, updating menus, handling sales data, and interacting with the database.DailySoldStats, FollowShopRequest, LoginRequest, LoginResponse, etc., define data structures and responses for specific use cases like user authentication, shop statistics, and shop-following requests.
    Interfaces like ShopOperations and ShopRoutesInterface define the core functionality and routes for shops, which are implemented by the Shop class.

### models namespace: 
    Defines the domain model for the platform, including entities like Shop, Item, Account, Reviews, and more. These models represent the data structures used across the platform, such as items with prices and availability, shop details, user accounts, etc.

### repository namespace:
    Defines the database-related operations. The DataBase class manages CRUD operations for shops, items, and user accounts. The ShopRepository and UserRepository interfaces abstract these database operations.

### initializer namespace:
    Contains the Config class, which holds configuration values such as database credentials, server port, JWT secrets, and SMTP settings.

### routes namespace:
    Defines HTTP routes related to user and shop management. For example, ShopRoutes and UserRoute classes set up routes for shop-related actions and user authentication.

### scheduleUpdates namespace:
    Defines classes related to scheduling tasks and cron jobs. The CustomCronJob and UpdateDB classes suggest background tasks for periodic updates, such as scraping new data, updating sold items, or handling item updates.



## Design Considerations
Modularity: The system is broken into distinct namespaces (collector, controllers, repository, etc.), making it easier to manage and extend different parts of the application independently.
Separation of Concerns: Each class and interface has a clear responsibility (e.g., database operations, user management, sales statistics).




