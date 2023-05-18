global:
  scrape_interval: 1m
  scrape_timeout: 10s
scrape_configs:
  - job_name: kubernetes-service-endpoints
    sample_limit: 10000
    kubernetes_sd_configs:
    - role: endpoints
    relabel_configs:
    - action: keep
      regex: true
      source_labels:
      - __meta_kubernetes_service_annotation_prometheus_io_scrape
    - action: replace
      regex: (https?)
      source_labels:
      - __meta_kubernetes_service_annotation_prometheus_io_scheme
      target_label: __scheme__
    - action: replace
      regex: (.+)
      source_labels:
      - __meta_kubernetes_service_annotation_prometheus_io_path
      target_label: __metrics_path__
    - action: replace
      regex: ([^:]+)(?::\d+)?;(\d+)
      replacement: $1:$2
      source_labels:
      - __address__
      - __meta_kubernetes_service_annotation_prometheus_io_port
      target_label: __address__
    - action: labelmap
      regex: __meta_kubernetes_service_label_(.+)
    - action: replace
      source_labels:
      - __meta_kubernetes_namespace
      target_label: Namespace
    - action: replace
      source_labels:
      - __meta_kubernetes_service_name
      target_label: Service
    - action: replace
      source_labels:
      - __meta_kubernetes_pod_node_name
      target_label: kubernetes_node
    - action: replace
      source_labels:
      - __meta_kubernetes_pod_name
      target_label: pod_name
    - action: replace
      source_labels:
      - __meta_kubernetes_pod_container_name
      target_label: container_name
    metric_relabel_configs:
    - source_labels: [__name__]
      regex: 'go_gc_duration_seconds.*'
      action: drop
    - source_labels: [__name__, proxy]
      regex: "haproxy_frontend.+;(.+)"
      target_label: frontend
      replacement: "$1"
    - source_labels: [__name__, proxy]
      regex: "haproxy_server.+;(.+)"
      target_label: backend
      replacement: "$1"
    - source_labels: [__name__, proxy]
      regex: "haproxy_backend.+;(.+)"
      target_label: backend
      replacement: "$1"
    - regex: proxy
      action: labeldrop