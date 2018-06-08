curl -L https://download.docker.com/linux/debian/dists/$(cat /etc/debian_version | cut -d'/' -f1)/pool/stable/amd64/docker-ce_18.03.1~ce-0~debian_amd64.deb -o /tmp/docker.deb && \
sudo dpkg -i /tmp/docker.deb && \
sudo curl -L https://github.com/docker/compose/releases/download/1.21.2/docker-compose-$(uname -s)-$(uname -m) -o /usr/local/bin/docker-compose && \
sudo chmod +x /usr/local/bin/docker-compose && \
sudo docker-compose up -d