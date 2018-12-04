# -*- mode: ruby -*-
# vi: set ft=ruby :

# For a complete reference, please see the online documentation at
# https://docs.vagrantup.com.

# Every Vagrant development environment requires a box. You can search for
# boxes at https://vagrantcloud.com/search.

Vagrant.configure("2") do |config|
  # The most common configuration options are documented and commented below.

  config.vm.box = "bento/ubuntu-18.04"
  config.ssh.insert_key = false

  config.vm.hostname = "loca.local"
  config.vm.post_up_message = "Run 'vagrant ssh' and do what it says "
  config.vm.network "forwarded_port", guest: 80, host: 8080
  config.vm.network "public_network"

  # Make sure the local repo is there
  config.vm.synced_folder "infra/plays/", "/srv/otto/plays/"

  # virtualbox is the "provider"
  config.vm.provider "virtualbox" do |vb|
    # Customize the amount of memory on the VM:
    vb.memory = "1024"  # make this smaller for production
    vb.linked_clone = true
  end

  # ansible is the "provisioner"
  # config.vm.provision "ansible" do |ansible|
  config.vm.provision "ansible_local" do |ansible|
    ansible.playbook = "site.yml"
    ansible.provisioning_path = "/srv/otto/infra/plays"
    ansible.install = true
    ansible.install_mode = "pip"
  end

  # # otto is our application
  # config.vm.define "otto" do |app|
  #   app.vm.hostname = "otto.test"
  #   app.vm.network :private_network, ip: "10.24.13.11"
  # end

  # otto is our application
  config.vm.define "o02" do |app|
    app.vm.hostname = "o02.test"
    app.vm.network :private_network, ip: "10.24.13.2"
  end

  ## Additional hosts can be defined here ...
  # config.vm.define "app2" do |app|
  #   app.vm.hostname = "o1-app2.test"
  #   app.vm.network :private_network, ip: "192.168.60.5"
  # end

  # config.vm.define "haproxy" do |app|
  #   app.vm.hostname = "o1-haproxy.test"
  #   app.vm.network :private_network, ip: "192.168.60.6"
  # end

end
