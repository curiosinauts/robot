apiVersion: apps/v1
kind: Deployment
metadata:
  name: robot
  labels:
    app: robot
spec:
  replicas: 1
  selector:
    matchLabels:
      app: robot
  template:
    metadata:
      labels:
        app: robot
    spec:
      containers:
        - env:
          - name: SLACK_BOT_USER_OAUTH_ACCESS_TOKEN
            valueFrom:
              secretKeyRef:
                key: slack-bot-user-oauth-access-token
                name: robot-secrets
          - name: SLACK_SIGNING_SECRET
            valueFrom:
              secretKeyRef:
                key: slack-signing-secret
                name: robot-secrets
          - name: PLATFORMCTL_DATABASE_CONN
            valueFrom:
              secretKeyRef:
                key: platformctl-database-conn
                name: robot-secrets
          - name: PLATFORMCTL_GITEA_ACCESS_TOKEN
            valueFrom:
              secretKeyRef:
                key: platformctl-gitea-access-token
                name: robot-secrets
          - name: PLATFORMCTL_GITEA_URL
            valueFrom:
              secretKeyRef:
                key: platformctl-gitea-url
                name: robot-secrets
          - name: PLATFORMCTL_JENKINS_API_KEY
            valueFrom:
              secretKeyRef:
                key: platformctl-jenkins-api-key
                name: robot-secrets
          - name: PLATFORMCTL_JENKINS_URL
            valueFrom:
              secretKeyRef:
                key: platformctl-jenkins-url
                name: robot-secrets
          - name: PLATFORMCTL_DOCKER_REGISTRY_URL
            valueFrom:
              secretKeyRef:
                key: platformctl-docker-registry-url
                name: robot-secrets
          - name: PLATFORMCTL_DOCKER_REGISTRY_HOST
            valueFrom:
              secretKeyRef:
                key: platformctl-docker-registry-host
                name: robot-secrets
          - name: PLATFORMCTL_PIHOLE_HOST
            valueFrom:
              secretKeyRef:
                key: platformctl-pihole-host
                name: robot-secrets
          - name: PLATFORMCTL_SHARED_DATABASE_HOST
            valueFrom:
              secretKeyRef:
                key: platformctl-shared-database-host
                name: robot-secrets
          - name: PLATFORMCTL_SHARED_DATABASE_NAME
            valueFrom:
              secretKeyRef:
                key: platformctl-shared-database-name
                name: robot-secrets
          - name: PLATFORMCTL_GRAFANA_API_KEY
            valueFrom:
              secretKeyRef:
                key: platformctl-grafana-api-key
                name: robot-secrets
          name: robot
          image: docker-registry.curiosityworks.org/curiosinauts/robot:__tag__
          ports:
            - containerPort: 3000
      dnsPolicy: Default

---
apiVersion: v1
kind: Service
metadata:
  name: robot
spec:
  selector:
    app: robot
  ports:
    - protocol: TCP
      port: 80
      targetPort: 3000

---
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: robot
  annotations:
    kubernetes.io/ingress.class: "traefik"
spec:
  rules:
    - host: robot.curiosityworks.org
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: robot
                port:
                  number: 80