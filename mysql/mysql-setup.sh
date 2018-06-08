sudo apt install -y mysql-server-5.7 && \
sudo mysql_secure_installation && \
echo "init.sql"
mysql -h 127.0.0.1 -u root -p < init.sql

sudo sed -i 's/127\.0\.0\.1/0\.0\.0\.0/g' /etc/mysql/my.cnf
echo "some script from the interwebs"
mysql -uroot -p -e 'USE mysql; UPDATE `user` SET `Host`="%" WHERE `User`="root" AND `Host`="localhost"; DELETE FROM `user` WHERE `Host` != "%" AND `User`="root"; FLUSH PRIVILEGES;'

sudo service mysql restart