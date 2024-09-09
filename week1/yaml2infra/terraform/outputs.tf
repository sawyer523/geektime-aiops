output "kube_config" {
  description = "kubeconfig"
  value       = "${path.module}/config.yaml"
}