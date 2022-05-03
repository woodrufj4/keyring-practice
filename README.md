# Keyring Practice

This is just a learning CLI that emulates some of [HashiCorp Vault's](https://github.com/hashicorp/vault) encrypt / decrypt funcationality.

The main learning objectives:
- [x] Learn how to use the underlying [AEAD](https://pkg.go.dev/crypto/cipher#AEAD) library
- [ ] Learn how to leverage a keyring for the secret key use and rotation
- [ ] Learn how to securely persist secrets, and recover persisted secrets from the event of disaster recovery
- [ ] Learn the meaning and practical use of a [nonce](https://en.wikipedia.org/wiki/Cryptographic_nonce) and [authenticated encryption](https://en.wikipedia.org/wiki/Authenticated_encryption)
