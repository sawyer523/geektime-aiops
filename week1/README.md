1. 实践 Terraform，开通腾讯云虚拟机，并安装 Docker。
```shell
❯ export TF_VAR_secret_id=xxx
❯ export export TF_VAR_secret_key=xxx
❯ terraform init
❯ terraform plan
❯ terraform apply -auto-approve
tencentcloud_security_group.default: Refreshing state... [id=sg-odyhrziz]
data.tencentcloud_images.default: Reading...
data.tencentcloud_availability_zones_by_product.default: Reading...
data.tencentcloud_instance_types.default: Reading...
data.tencentcloud_availability_zones_by_product.default: Read complete after 0s [id=1851109860]
tencentcloud_security_group_lite_rule.default: Refreshing state... [id=sg-odyhrziz]
data.tencentcloud_images.default: Read complete after 1s [id=572099355]
data.tencentcloud_instance_types.default: Read complete after 1s [id=101746864]

Terraform used the selected providers to generate the following execution plan. Resource actions are
indicated with the following symbols:
+ create

Terraform will perform the following actions:

# tencentcloud_instance.web[0] will be created
+ resource "tencentcloud_instance" "web" {
  + allocate_public_ip                      = true
  + availability_zone                       = "ap-hongkong-2"
  + create_time                             = (known after apply)
  + disable_api_termination                 = false
  + disable_monitor_service                 = false
  + disable_security_service                = false
  + expired_time                            = (known after apply)
  + force_delete                            = false
  + id                                      = (known after apply)
  + image_id                                = "img-mmytdhbn"
  + instance_charge_type                    = "SPOTPAID"
  + instance_charge_type_prepaid_renew_flag = (known after apply)
  + instance_name                           = "web server"
  + instance_status                         = (known after apply)
  + instance_type                           = "SA5.MEDIUM4"
  + internet_charge_type                    = (known after apply)
  + internet_max_bandwidth_out              = 100
  + key_ids                                 = (known after apply)
  + key_name                                = (known after apply)
  + orderly_security_groups                 = [
      + "sg-odyhrziz",
    ]
  + password                                = (sensitive value)
  + private_ip                              = (known after apply)
  + project_id                              = 0
  + public_ip                               = (known after apply)
  + running_flag                            = true
  + security_groups                         = (known after apply)
  + subnet_id                               = (known after apply)
  + system_disk_id                          = (known after apply)
  + system_disk_size                        = 50
  + system_disk_type                        = "CLOUD_BSSD"
  + user_data                               = "IyEvYmluL2Jhc2gKIyDmm7TmlrDljIXliJfooagKYXB0LWdldCB1cGRhdGUKCiMg5a6J6KOF5b+F6KaB55qE5L6d6LWWCmFwdC1nZXQgaW5zdGFsbCAteSBhcHQtdHJhbnNwb3J0LWh0dHBzIGNhLWNlcnRpZmljYXRlcyBjdXJsIHNvZnR3YXJlLXByb3BlcnRpZXMtY29tbW9uCgojIOa3u+WKoCBEb2NrZXIg55qE5a6Y5pa5IEdQRyDlr4bpkqUKY3VybCAtZnNTTCBodHRwczovL2Rvd25sb2FkLmRvY2tlci5jb20vbGludXgvdWJ1bnR1L2dwZyB8IGFwdC1rZXkgYWRkIC0KCiMg6K6+572uIERvY2tlciDnmoTnqLPlrprniYjku5PlupMKYWRkLWFwdC1yZXBvc2l0b3J5ICJkZWIgW2FyY2g9YW1kNjRdIGh0dHBzOi8vZG93bmxvYWQuZG9ja2VyLmNvbS9saW51eC91YnVudHUgJChsc2JfcmVsZWFzZSAtY3MpIHN0YWJsZSIKCiMg5YaN5qyh5pu05paw5YyF5YiX6KGoCmFwdC1nZXQgdXBkYXRlCgojIOWuieijhSBEb2NrZXIKYXB0LWdldCBpbnN0YWxsIC15IGRvY2tlci1jZQoKIyDlkK/liqggRG9ja2VyIOacjeWKoQpzeXN0ZW1jdGwgc3RhcnQgZG9ja2VyCgojIOiuvue9riBEb2NrZXIg5byA5py66Ieq5ZCvCnN5c3RlbWN0bCBlbmFibGUgZG9ja2VyCgojIOi+k+WHuiBEb2NrZXIg54mI5pys77yM5Lul6aqM6K+B5a6J6KOFCmRvY2tlciAtLXZlcnNpb24="
  + vpc_id                                  = (known after apply)

  + data_disks (known after apply)
}

Plan: 1 to add, 0 to change, 0 to destroy.

Changes to Outputs:
~ private_ip = "172.19.0.13" -> (known after apply)
~ public_ip  = "43.135.76.163" -> (known after apply)
tencentcloud_instance.web[0]: Creating...
tencentcloud_instance.web[0]: Still creating... [10s elapsed]
tencentcloud_instance.web[0]: Still creating... [20s elapsed]
tencentcloud_instance.web[0]: Still creating... [30s elapsed]
tencentcloud_instance.web[0]: Creation complete after 33s [id=ins-d5ve72gy]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

Outputs:

private_ip = "172.19.0.17"
public_ip = "43.154.142.128"

❯ ssh ubuntu@43.154.142.128
ubuntu@VM-0-17-ubuntu:~$ docker version
Client: Docker Engine - Community
 Version:           27.2.0
 API version:       1.47
 Go version:        go1.21.13
 Git commit:        3ab4256
 Built:             Tue Aug 27 14:15:15 2024
 OS/Arch:           linux/amd64
 Context:           default
permission denied while trying to connect to the Docker daemon socket at unix:///var/run/docker.sock: Get "http://%2Fvar%2Frun%2Fdocker.sock/v1.47/version": dial unix /var/run/docker.sock: connect: permission denied
```

1. 使用 YAML to Infra 模式创建云 Redis 数据库。
    1. 使用 terraform 初始化环境
    ```shell
    ❯ export TF_VAR_secret_id=xxx
    ❯ export export TF_VAR_secret_key=xxx
    ❯ terraform init
    ❯ terraform plan
    ❯ terraform apply -auto-approve
    ❯ export KUBECONFIG=./config.yaml
    ❯ k get ns
    NAME                STATUS   AGE
    argocd              Active   5m57s
    crossplane-system   Active   6m1s
    default             Active   6m19s
    kube-node-lease     Active   6m19s
    kube-public         Active   6m19s
    kube-system         Active   6m19s
    ```
    2. 创建 Crossplane provider 和 provider config
    ```shell
    ❯ k -n crossplane-system apply -f ../yaml/tf-provider.yaml
    provider.pkg.crossplane.io/provider-terraform created
    ❯ k -n crossplane-system get providers
    NAME                 INSTALLED   HEALTHY   PACKAGE                                              AGE
    provider-terraform   True        True      xpkg.upbound.io/upbound/provider-terraform:v0.18.0   18s
    ❯ k -n crossplane-system apply -f ../yaml/tf-provider-config.yaml
    providerconfig.tf.upbound.io/default created
    ```
    3. 