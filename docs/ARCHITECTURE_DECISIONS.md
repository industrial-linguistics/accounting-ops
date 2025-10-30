# Architecture Decision Records

This document tracks key architectural decisions made during the development of the Accounting Operations Toolkit.

## ADR-001: Separate Broker Instances for Development and Production

**Date:** 2025-10-30

**Status:** Accepted

### Context

The OAuth broker needs to support both sandbox/development and production environments for multiple providers (QuickBooks, Xero, Deputy). Each provider uses different OAuth credentials and API base URLs depending on the environment.

We evaluated two approaches:
1. **Per-request environment selection**: Single broker instance that accepts an `environment` parameter and selects credentials/URLs dynamically
2. **Separate broker instances**: Deploy distinct brokers for development and production, each with their own configuration

### Decision

We will use **separate broker instances**:
- **Production**: `https://auth.industrial-linguistics.com/cgi-bin/broker`
- **Development**: `https://auth-dev.industrial-linguistics.com/cgi-bin/broker`

The CLI will support the `ACCOUNTING_OPS_BROKER` environment variable to override the default broker URL.

### Rationale

**Advantages:**
- **Simpler broker logic**: Each broker serves one environment, no dynamic credential selection
- **Configuration isolation**: Development and production credentials completely separated
- **Staging/testing workflow**: Provides a natural staging area for broker updates before production deployment
- **Security**: Reduced risk of accidentally using production credentials in development
- **Debugging**: Easier to troubleshoot when each environment is isolated

**Disadvantages:**
- **Infrastructure overhead**: Requires two broker deployments and DNS entries
- **Credential duplication**: Need to maintain separate broker.env files for each environment

The simplicity and isolation benefits outweigh the infrastructure overhead, especially given the infrequent credential changes and the value of having a staging environment.

### Implementation

**Broker Configuration:**
- Each broker instance loads its own `broker.env` file at startup
- Development broker.env contains sandbox/development credentials
- Production broker.env contains production credentials

**CLI:**
```bash
# Default: production
acct connect qbo --profile mycompany

# Development: override via environment variable
export ACCOUNTING_OPS_BROKER=https://auth-dev.industrial-linguistics.com/cgi-bin/broker
acct connect qbo --profile mycompany

# Or via flag
acct connect qbo --profile mycompany --broker https://auth-dev.industrial-linguistics.com/cgi-bin/broker
```

**DNS Configuration:**
- `auth.industrial-linguistics.com` → production OpenBSD server
- `auth-dev.industrial-linguistics.com` → development OpenBSD server (or same server, different virtual host)

**Deployment:**
- GitHub Actions workflow can deploy to both environments
- Development deployment can be automatic on push to `develop` branch
- Production deployment can require manual approval or be triggered by tags

### Alternatives Considered

**Per-request environment selection:**
- CLI would send `{"provider": "qbo", "environment": "sandbox"}` to broker
- Broker would select credentials dynamically: `QBO_SANDBOX_CLIENT_ID` vs `QBO_CLIENT_ID`
- Single broker URL for all environments

This was rejected due to increased complexity in credential management and potential for environment confusion. The separate-instance approach provides clearer separation and better aligns with standard staging/production workflows.

### Consequences

- DNS entry for `auth-dev.industrial-linguistics.com` must be created
- Development broker deployment process must be established
- Documentation must clearly indicate which broker URL to use for development vs production
- Environment variable `ACCOUNTING_OPS_BROKER` must be added to CLI
- Profile credentials stored in OS keychain will be implicitly tied to the broker environment used during `acct connect`

### Related Documents

- [docs/BROKER_ENV_TEMPLATE.md](BROKER_ENV_TEMPLATE.md) - Environment variable configuration
- [docs/QUICKBOOKS_DEPLOYMENT_CHECKLIST.md](QUICKBOOKS_DEPLOYMENT_CHECKLIST.md) - Deployment procedures
- [CLAUDE.md](../CLAUDE.md) - Project architecture overview
