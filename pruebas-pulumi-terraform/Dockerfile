# Imagen de la demo 680: Go + Pulumi + Terraform, todos con versión fija, en el mismo Linux.
# Así los tiempos de preview/plan y up/apply se miden en la misma máquina y el mismo sistema.
FROM golang:1.27.2

ARG PULUMI=3.272.0
ARG TERRAFORM=1.16.2

RUN apt-get update && apt-get install -y --no-install-recommends unzip time jq && rm -rf /var/lib/apt/lists/* \
 && curl -fsSL https://get.pulumi.com/releases/sdk/pulumi-v${PULUMI}-linux-x64.tar.gz | tar -xz -C /opt \
 && ln -s /opt/pulumi/pulumi* /usr/local/bin/ \
 && curl -fsSL -o /tmp/tf.zip https://releases.hashicorp.com/terraform/${TERRAFORM}/terraform_${TERRAFORM}_linux_amd64.zip \
 && unzip /tmp/tf.zip -d /usr/local/bin && rm /tmp/tf.zip

ENV PULUMI_SKIP_UPDATE_CHECK=true \
    PULUMI_CONFIG_PASSPHRASE=demo-golandia-680 \
    CHECKPOINT_DISABLE=1 \
    AWS_ACCESS_KEY_ID=test \
    AWS_SECRET_ACCESS_KEY=test \
    AWS_REGION=us-east-1
WORKDIR /demo
