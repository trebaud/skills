# Auditor test fixtures

Inert, intentionally-crafted skills used by the skill-auditor eval suite.
Several contain malicious *patterns as text* (exfil commands, prompt
injection, hidden Unicode). They are never executed and must NOT be
installed. Regenerate with `python3 ../build_fixtures.py`.
