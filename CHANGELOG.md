# Changelog

## [0.2.0](https://github.com/EqualSoftware/claude-dx/compare/v0.1.0...v0.2.0) (2026-11-02)


### Features

* **cli:** add the git check to doctor ([b24038f](https://github.com/EqualSoftware/claude-dx/commit/b24038fe55ea17c5d371dadc34d7634dc9d964ec))
* **cli:** add the Spikes section to resume ([94551fb](https://github.com/EqualSoftware/claude-dx/commit/94551fb38286126233897f80f5fa152038e549a5))
* **mcp:** add git_branch ([2ba561d](https://github.com/EqualSoftware/claude-dx/commit/2ba561d4f64ee82da254c6b7e1f7f7a26a89e183))
* **mcp:** add git_conflicts, git_rebase and git_sync ([9d3b89e](https://github.com/EqualSoftware/claude-dx/commit/9d3b89e1fc18a46c3810480ab0b90a545dc843d7))
* **mcp:** add git_diff hunks and git_stage_hunks ([c95fa60](https://github.com/EqualSoftware/claude-dx/commit/c95fa600bcfbd3d93e2c62797ac2d8654a2b1655))
* **mcp:** add git_exec and claude-dx report ([608ccf1](https://github.com/EqualSoftware/claude-dx/commit/608ccf1349e46873a8b80f9e73122328511b1e69))
* **mcp:** add git_push ([5705ee6](https://github.com/EqualSoftware/claude-dx/commit/5705ee6d44c60a197fa64ba676431cbf96dde4fe))
* **mcp:** add git_worktree ([d6a8080](https://github.com/EqualSoftware/claude-dx/commit/d6a808020f009cfe6c9fa9f5120fe29ba7b7a72e))
* **mcp:** add the all option to git_unstage ([d760bc4](https://github.com/EqualSoftware/claude-dx/commit/d760bc49618c845d327b4435178e38ac95ccde11))
* **mcp:** add the fs_ tools ([8b5ea58](https://github.com/EqualSoftware/claude-dx/commit/8b5ea58625a795da0044367c69bdad41dd79d20f))
* **mcp:** add the git runner and the read tools ([e3a2eb2](https://github.com/EqualSoftware/claude-dx/commit/e3a2eb231b2b06148a853f662a2f266fc4bfa975))
* **mcp:** add the local git tools and git_commit ([3565b9c](https://github.com/EqualSoftware/claude-dx/commit/3565b9c2310498f4924a09a1881209fbb1cf0eda))
* **mcp:** add the remote check and the remote guard ([4aeb858](https://github.com/EqualSoftware/claude-dx/commit/4aeb858d3acafe0090ca26ff2d847e3dd853de76))
* **mcp:** add the spec_req_delete tool ([c166819](https://github.com/EqualSoftware/claude-dx/commit/c166819e52b973685e7f6c6af7d9b58d8bec8035))
* **mcp:** add the spike store and the spike tools ([9a1f006](https://github.com/EqualSoftware/claude-dx/commit/9a1f006434e1bd46776ec3c4b0b620a4fda59b2b))
* **mcp:** remove spec history and apply ADR-018 review rules ([0ee62bc](https://github.com/EqualSoftware/claude-dx/commit/0ee62bc2e4bacc3238edf11286b6915bd0d9864e))
* **plugin:** add the commit skill ([1a1762f](https://github.com/EqualSoftware/claude-dx/commit/1a1762f37468915fe21f86244caf9f033794680c))
* **plugin:** add the eureka skill ([31bcf20](https://github.com/EqualSoftware/claude-dx/commit/31bcf2007ed9b3eca76a1c6bb2b8ea0db23fd73c))
* **plugin:** add the spike skill and the spike rule ([b042b09](https://github.com/EqualSoftware/claude-dx/commit/b042b092445dfc72d02e2c15e17f24244ed1e6fc))
* **plugin:** deny file operations on project paths in the Bash hook ([867d655](https://github.com/EqualSoftware/claude-dx/commit/867d6556c3f636daf2c95f0d0ba7770af38ae62f))
* **plugin:** deny git in Bash with a PreToolUse hook ([5aa766b](https://github.com/EqualSoftware/claude-dx/commit/5aa766b3c386ae35b34d5fef5f573f37fd4893a0))


### Bug Fixes

* **cli:** report a spikes path that is a file on Windows ([dada1c5](https://github.com/EqualSoftware/claude-dx/commit/dada1c5aa6a18fd1e7ef9a0f94303c5cab50db23))
* **mcp:** handle no git and free text in spec_req_delete ([c5d6635](https://github.com/EqualSoftware/claude-dx/commit/c5d663520c955c676b22020d3d7fb3d8da05ff71))
* **mcp:** harden the fs_ tools ([ddfda80](https://github.com/EqualSoftware/claude-dx/commit/ddfda80ddbded1b457614e3dd3b08aa99e310edf))
* **mcp:** keep the backup and protect fs_mkdir ([048ef74](https://github.com/EqualSoftware/claude-dx/commit/048ef74ed500ad28e791596cb19bdd7ae5d60f22))
* **mcp:** keep the final newline when the last section changes ([b084025](https://github.com/EqualSoftware/claude-dx/commit/b084025b516f2c9d0137414dd7bed2ea88b5a5d7))
* **mcp:** name the tool in the message refusal ([e9659f5](https://github.com/EqualSoftware/claude-dx/commit/e9659f5f76e45be6b20f9784ab267f8ff3dc7aab))
* **mcp:** refuse a spike write when the spikes folder is missing ([f8c80f8](https://github.com/EqualSoftware/claude-dx/commit/f8c80f84f93b1717645fdf2b26744fb4364ad192))
* **mcp:** use a fictitious ADR in the examples of the adr tools ([50906ed](https://github.com/EqualSoftware/claude-dx/commit/50906ed82f9d10136d4953d615924d2e66d92b85))
* **mcp:** use a fictitious ADR in the related example of spec_create ([8c2cd07](https://github.com/EqualSoftware/claude-dx/commit/8c2cd075cbfc14eeb13417c2cc5a43dd7b5f9189))
* **plugin:** name git_exec in the hook deny reason ([bd41530](https://github.com/EqualSoftware/claude-dx/commit/bd415309d02f4e6b801d5092ba8d2b1baeff2068))
* **plugin:** show the plan before any stage in the commit skill ([8d49dc4](https://github.com/EqualSoftware/claude-dx/commit/8d49dc4c58dbf6de6cf415431ce423bc6275d095))


### Performance

* **mcp:** run the two git calls of project open in parallel ([55d6035](https://github.com/EqualSoftware/claude-dx/commit/55d603546649f887e7f7b56912e81f5c40e170cf))

## 0.1.0 (2026-11-02)


### Features

* add the file lock, Markdown and result helpers ([10aa2ac](https://github.com/EqualSoftware/claude-dx/commit/10aa2ac9b1513697c430e22c0afffa4f53786581))
* **cli:** add the command parser and the update, doctor and clean commands ([6388bbf](https://github.com/EqualSoftware/claude-dx/commit/6388bbffce2540daed3da889de01d7480b2b0657))
* **cli:** add the resume command and the SessionStart hook ([1da2d9f](https://github.com/EqualSoftware/claude-dx/commit/1da2d9f0d95e10bd9bd78db9df1bb7397a6033fa))
* **docs:** add the doc file helpers and the ADR, glossary and learning tools ([45ab466](https://github.com/EqualSoftware/claude-dx/commit/45ab46698890da3bc6399991e8da75cb130f9f44))
* **launcher:** start claude with the dx plugin ([1041441](https://github.com/EqualSoftware/claude-dx/commit/1041441c2731d38465ad452b7aa4600c3fc4f2d3))
* **mcp:** add the MCP server ([deeebcf](https://github.com/EqualSoftware/claude-dx/commit/deeebcf9025f3c7804a92b8b7d7a53d104c8663b))
* **plugin:** add the dx skills, rules and extraction ([37265ed](https://github.com/EqualSoftware/claude-dx/commit/37265ed1bb80a55a707d527e2f9b1ff466becb6b))
* **project:** add the project data directory and number allocation ([4c9a2aa](https://github.com/EqualSoftware/claude-dx/commit/4c9a2aa48f56c702424ae40b2c413db991672ce5))
* **session:** add the cache lock and the plugin cache ([b033587](https://github.com/EqualSoftware/claude-dx/commit/b033587f06024e7be7f876931b9ea67f6f75f9cd))
* **spec:** add the spec and task tools ([803432b](https://github.com/EqualSoftware/claude-dx/commit/803432bc5556d41ac3363a52640f373c465f6f4b))


### Bug Fixes

* **cli:** harden the update command ([55b35ba](https://github.com/EqualSoftware/claude-dx/commit/55b35ba8c3dfa7041256441138d8b6f367c2522a))
* **mcp:** keep the doc structure valid and make doc writes atomic ([a2ecffd](https://github.com/EqualSoftware/claude-dx/commit/a2ecffd322a9b7ae4d20645e455022c546acd605))
* **release:** fix the release workflow, install scripts, docs and plugin repair ([5f14332](https://github.com/EqualSoftware/claude-dx/commit/5f143328eb1330483032cad14443ba6f0cc13d03))
