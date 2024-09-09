variable "secret_id" {
  default = "Your Access ID"
}

variable "secret_key" {
  default = "Your Access Key"
}

variable "region" {
    description = "The region in which the resources will be created"
    default     = "ap-hongkong"
}

variable "password" {
    description = "The password for the CVM instance"
    default     = "Password1234!"
}