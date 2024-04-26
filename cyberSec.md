# Threat Model

This document provides an in-depth analysis of potential security threats and the corresponding measures implemented or planned to safeguard our Golang backend system.

Threat Model:
Recognize various threat vectors and their implications on the system's security. These include authentication vulnerabilities, authorization breaches, XSS attacks, CSRF exploits, server-side vulnerabilities.

## Security Measures

### Authentication

* Utilization of bcrypt for password hashing to enhance security during user authentication.
* Consideration of adopting scrypt for even stronger password hashing.
* Deployment of JSON Web Tokens (JWTs) for user authentication, with expiration and refresh mechanisms.
* Deployment of blacklisting compromised JWTs to mitigate token-based attacks.

### Authorization:

* Utilization of middleware to enforce authentication requirements for accessing private user data.

### Front-End Application (React)

* Deployment of HTTP-only cookies for JWT storage to mitigate client-side token theft.
* Utilization of built-in React-js measures for XSS prevention.
* Planned Implementation of CSRF tokens to prevent Cross-Site Request Forgery attacks.

### Environment Security

* Protection of environment secrets from exposure in GitHub repository.
* Planned implementation of Vault service for secure storage and deployment of environment secrets.

### PostgreSQL Database Security:

**Network Isolation**

* Implementation of network segmentation to isolate the PostgreSQL database from unauthorized access.
* Utilization of firewalls and network security groups to restrict incoming connections to trusted sources only.

**Encryption**

* Implementation of SSL/TLS encryption for secure communication between client applications and the database server.

**Access Control**

* Implementation of role-based access control (RBAC) within PostgreSQL to define granular permissions for database objects.
* Utilization of whitelisted users and IPs to control access to the database, allowing only authorized users and trusted IP addresses to connect.
