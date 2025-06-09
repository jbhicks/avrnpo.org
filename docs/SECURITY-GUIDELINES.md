# 🚨 CRITICAL SECURITY GUIDELINES 🚨

## NEVER EXPOSE SENSITIVE DATA

### Absolute Prohibitions

**NEVER** include real sensitive data in:
- Documentation files (.md, .txt, .rst, etc.)
- Code comments or examples
- Test files or fixtures
- Log files or output
- Git commits or version control
- Any file that could be shared or published

### What Constitutes Sensitive Data

- API keys, tokens, secrets
- Database passwords or connection strings
- Private keys or certificates
- User passwords or hashes
- Session tokens or JWTs
- OAuth client secrets
- Webhook secrets
- Any production credentials

### Required Practices

1. **Use Placeholders Only**
   ```
   ✅ CORRECT: HELCIM_PRIVATE_API_KEY=your_helcim_api_key_here
   ❌ NEVER: HELCIM_PRIVATE_API_KEY=js_1234567890123456789
   ```

2. **Environment Variables**
   - Store all secrets in `.env` files (never committed)
   - Use `.env.example` with placeholder values
   - Reference environment variables in documentation, never actual values

3. **Documentation Examples**
   - Always use fake/placeholder data
   - Use patterns like `your_key_here`, `REPLACE_WITH_ACTUAL_VALUE`
   - Never copy-paste real configuration

4. **Immediate Response to Exposure**
   - If real credentials are found in any file, immediately replace with placeholders
   - Rotate/revoke the exposed credentials if they were real
   - Review all related files for additional exposure

### Security Review Checklist

Before creating or editing any file:
- [ ] Does this contain any real API keys, tokens, or passwords?
- [ ] Are all examples using placeholder values?
- [ ] Would this be safe to share publicly?
- [ ] Are sensitive values properly stored in environment variables?

## Implementation Notes

- This applies to ALL files in the project, no exceptions
- Documentation must be safe for public viewing
- Test files should use mock/fake credentials only
- Examples should demonstrate structure, not real values
