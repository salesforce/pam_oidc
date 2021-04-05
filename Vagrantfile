# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/focal64"

  config.vm.network "private_network", ip: "100.121.23.200"

  config.vm.provision "shell", inline: <<-SHELL
    export DEBIAN_FRONTEND=noninteractive

    apt-get update
    apt-get install -y build-essential libpam0g-dev gnupg2 wget

    if ! dpkg -l percona-release &>/dev/null; then
      wget https://repo.percona.com/apt/percona-release_latest.$(lsb_release -sc)_all.deb
      dpkg -i percona-release_latest.$(lsb_release -sc)_all.deb

      apt-get update
    fi

    if ! dpkg -l percona-server-server-5.7 &>/dev/null; then
      apt-get install -y percona-server-server-5.7
    fi

    if ! snap list go &>/dev/null; then
      snap install --classic go
    fi

    cd /vagrant
    make

    if [ ! -f "/etc/pam.d/mysqld" ]; then
      cat <<EOF >/etc/pam.d/mysqld
auth required /vagrant/pam_oidc.so issuer=https://accounts.google.com aud=32555940559.apps.googleusercontent.com [user_template={{.Extra.email}}]
auth required pam_warn.so

account required pam_permit.so
auth required pam_warn.so
EOF
    fi

   cat <<EOF | mysql -u root || true
INSTALL PLUGIN auth_pam SONAME 'auth_pam.so';
EOF

  systemctl restart mysql
  SHELL
end
