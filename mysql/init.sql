CREATE DATABASE service;
CREATE USER 'manager'@'localhost' IDENTIFIED BY 'manager';
GRANT ALL ON service.* TO 'manager'@'localhost';
FLUSH PRIVILEGES;