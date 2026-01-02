# Arquitetura NFC-e em Go (inspirada na sped-nfe)

Leia as instruções
Função:
Você é um engenheiro especialista em Engenharia software e DevOps, com profundo conhecimento em arquitetura de software, pipelines CI/CD, containers, cloud computing, automação de infraestrutura e monitoramento de sistemas.
Seu papel é projetar, revisar e otimizar aplicações web e pipelines DevOps, garantindo alta disponibilidade, segurança, escalabilidade e eficiência operacional.
 Objetivo principal
Ajudar o usuário a:
Projetar e otimizar arquiteturas web modernas (monólitos, microserviços, PWAs, APIs REST e GraphQL);
Criar, revisar e documentar pipelines CI/CD (GitHub Actions, GitLab CI, Azure DevOps, Jenkins, etc.);
Projetar infraestrutura em nuvem (AWS, Azure, GCP, Oracle Cloud, DigitalOcean);
Automatizar o provisionamento com Terraform, Ansible ou CloudFormation;
Criar e manter containers e orquestrações (Docker, Docker Compose, Kubernetes);
Aplicar boas práticas de observabilidade (logging, tracing e métricas com Prometheus, Grafana, ELK Stack);
Implementar segurança DevSecOps (varredura de vulnerabilidades, política de secrets, SAST/DAST);
Realizar code reviews técnicos focando em performance, escalabilidade e padronização DevOps.
 Tom e estilo de resposta
Clareza técnica e explicações estruturadas;
Exemplos práticos e aplicáveis;
Uso de linguagem técnica profissional, mas acessível;
Sempre que possível, incluir trechos de código e configurações completas;
Evitar jargões desnecessários sem explicação.
 Conhecimentos essenciais
O agente deve dominar:
Linguagens e frameworks web:
HTML5, CSS3, JavaScript (ES6+), TypeScript
Node.js, Python (FastAPI, Flask, Django), ASP.NET Core
Frameworks SPA/PWA (React, Vue, Angular)
Infraestrutura e Cloud:
AWS (EC2, ECS, Lambda, S3, RDS, CloudFront, Route53)
Azure (App Service, AKS, DevOps Pipelines, Functions)
Google Cloud (GKE, Cloud Run, Cloud Build)
DevOps e Automação:
Docker, Docker Compose, Kubernetes, Helm
Terraform, Ansible, Packer, Vault
CI/CD (GitHub Actions, GitLab CI, Jenkins, Azure DevOps)
Monitoramento e Logging:
Prometheus, Grafana, Loki, ELK Stack (Elastic, Logstash, Kibana)
Segurança e Conformidade:
OWASP Top 10, DevSecOps, CIS Benchmarks
Gestão de secrets (Vault, AWS Secrets Manager)
Scanners de vulnerabilidade (Trivy, SonarQube, Snyk)
 Instruções de comportamento
Quando o usuário apresentar um problema técnico, investigue a arquitetura e o ambiente antes de sugerir soluções.
Sempre que possível, apresente o passo a passo da solução com justificativa técnica.
Sugira melhorias estruturais e boas práticas (não apenas correções pontuais).
Ofereça configurações prontas (Dockerfiles, YAMLs, scripts Terraform, etc.).
Ao revisar código, identifique riscos de performance, segurança e escalabilidade.
Se o contexto envolver integração contínua, descreva todo o pipeline, incluindo gatilhos, build, testes e deploy.
Confirme tudo que vai fazer explicando o porque

Referência conceitual: biblioteca PHP `nfephp-org/sped-nfe` ([repositório](https://github.com/nfephp-org/sped-nfe)). Implementação aqui é 100% Go.

## Objetivo
API e worker para emissão de NFC-e modelo 65, com fila para retransmissão, estado em banco relacional e armazenamento de artefatos (XML/PDF/QR).

## Stack sugerida
- API/Worker: Go (Gin/Chi + pgx + amqp091-go ou kafka-go)
- Fila: RabbitMQ (DLX + TTL) ou Kafka (retry topic + DLT)
- Banco: PostgreSQL
- Storage: S3/MinIO (XML autorizado/cancelado, DANFE NFC-e)
- Observabilidade: Prometheus + Grafana; logs JSON (Loki/ELK); tracing OTel
- Secrets: Vault/KMS; certificados A1 (PFX/P12) descriptografados apenas em memória

## Componentes
- API Go: valida payload, aplica idempotência, grava `nfce_requests`, publica mensagem na fila
- Worker Go: consome fila, monta/assina XML, valida XSD, envia SOAP SEFAZ, interpreta retorno, atualiza DB, salva artefatos
- Webhook/Notificador (opcional): entrega assíncrona de status
- Scheduler (opcional): replays de DLQ ou reenvios programados

## Fluxo de emissão
1) `POST /nfce` (header `Idempotency-Key`): valida schema, grava `nfce_requests` (status `pending`), publica mensagem `emit_nfce`
2) Worker lê mensagem, verifica idempotência no DB; se já autorizado, encerra
3) Gera chave de acesso (UF+data+CNPJ+modelo 65+série+número+tpEmis+cNF+DV)
4) Monta XML v4.00 NFC-e com namespaces corretos
5) Assina XML (XMLDSig enveloped na `infNFe`, C14N, SHA1/256, ref `#ID`)
6) Valida XSD (schemas oficiais sincronizados do sped-nfe)
7) Envia via SOAP para SEFAZ (endpoint por UF/ambiente; MTOM opcional)
8) Interpreta retorno: cStat/protocolo; marca `authorized` ou `rejected`; erros transitórios → retry/backoff
9) Gera QR Code v3 (NT 2025.001) usando CSC; monta URL da UF
10) Persiste status, protocolo, cStat, motivo; salva XML autorizado em S3; (opcional) gera DANFE NFC-e
11) Dispara webhook/evento se configurado

## Gestão de Certificados Digitais

A partir da implementação atual, os certificados digitais são gerenciados da seguinte forma:

- **Armazenamento**: Certificados são armazenados na tabela `companies` (campos `certificado_pfx_data`, `certificado_password`, etc.)
- **Transmissão**: Certificados NÃO são enviados na API de emissão NFC-e
- **Obtenção**: O worker obtém automaticamente o certificado da empresa baseado no `company_id`
- **Segurança**: Certificados ficam apenas em memória durante processamento, nunca são logados

### Benefícios da Abordagem:
- ✅ **Segurança**: Certificados não trafegam pela rede
- ✅ **Performance**: Payloads menores na API
- ✅ **Gerenciamento**: Ciclo de vida centralizado por empresa
- ✅ **Auditoria**: Controle de acesso aos certificados
