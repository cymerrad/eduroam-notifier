CREATE DATABASE service;
CREATE USER 'django'@'localhost' IDENTIFIED BY 'django';
GRANT ALL ON service.* TO 'django'@'localhost';
FLUSH PRIVILEGES;