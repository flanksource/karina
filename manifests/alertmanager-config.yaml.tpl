global:
  # The smarthost and SMTP sender used for mail notifications.
  smtp_smarthost: 'mailvip1.discoveryhealth.co.za:25'
  smtp_from: 'alertmanager@discovery.co.za'
  resolve_timeout: "5m"


route:
  group_by: ['alertname']
  group_wait: 30s
  group_interval: 5m
  repeat_interval: 24h
  receiver: moshe


# Inhibition rules allow to mute a set of alerts given that another alert is
# firing.
# We use this to mute any warning-level notifications if the same alert is
# already critical.
inhibit_rules:
- source_match:
    severity: 'critical'
  target_match:
    severity: 'warning'
  # Apply inhibition if the alertname is the same.
  equal: ['alertname', 'cluster', 'service']

receivers:
- name: 'moshe'
  email_configs:
  - to: 'mosheai@discovery.co.za'
    require_tls: false
