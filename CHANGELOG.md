# Changelog

## 0.1.0 (2026-10-27)


### ⚠ BREAKING CHANGES

* **task:** the task tools take <spec>/<task-slug> IDs and new fields. The tools do not read docs/tasks.

### Features

* **adr:** add the ADR tools and the /dx:adr skill ([89cce61](https://github.com/EqualSoftware/claude-dx/commit/89cce61a6a783fee289c193561827bb88ba36dbb))
* **cli:** add the update, cache clean and doctor commands ([67aef6f](https://github.com/EqualSoftware/claude-dx/commit/67aef6f263aa3092eb179b65241b0e508338ae6d))
* **cli:** start claude with the embedded dx plugin ([7e985dc](https://github.com/EqualSoftware/claude-dx/commit/7e985dc819b18eae7de6b2208612158fca0641ff))
* **docs:** add a generated README index to each document folder ([5f958cf](https://github.com/EqualSoftware/claude-dx/commit/5f958cf7862bb29a9f7f5562bed9cf2744a79d26))
* **docs:** add the glossary_list tool ([bb19f07](https://github.com/EqualSoftware/claude-dx/commit/bb19f07a3c89264948dfe13b02665ef4917fbfa5))
* **docs:** add the learning and glossary tools and skills ([c23bc68](https://github.com/EqualSoftware/claude-dx/commit/c23bc68c7601d564c3bd139727f49d153b50ee40))
* **docs:** limit task and ADR titles to 40 characters ([cab6f35](https://github.com/EqualSoftware/claude-dx/commit/cab6f3516b1865030fbb315ee9fb4281f29b1030))
* **mcp:** add spec_update_fields ([357b990](https://github.com/EqualSoftware/claude-dx/commit/357b99046a4b6895191ce3fb7a854f9d5ea83923))
* **mcp:** add the dx MCP server and its shared packages ([70a74c1](https://github.com/EqualSoftware/claude-dx/commit/70a74c1ae592cd69f89bb3b81adc849de296e7b8))
* **mcp:** lock accepted ADRs after a push ([a8371ce](https://github.com/EqualSoftware/claude-dx/commit/a8371ce4227b3f622016f26f84ce171681a443a2))
* **plugin:** add the /dx:grill skill, the dx:scout agent and the shared rules ([81a3f5c](https://github.com/EqualSoftware/claude-dx/commit/81a3f5c0278fb390ab5cb1094baaf0bb9a486801))
* **plugin:** add the dx:spec skill and the spec rules ([d4a5437](https://github.com/EqualSoftware/claude-dx/commit/d4a5437932b12d3edc43a167e592dbbd7442f3bb))
* **plugin:** ask the user about choices found during work ([6922421](https://github.com/EqualSoftware/claude-dx/commit/6922421f5236c6d4d47bff38656101c95b62a4da))
* **project:** keep the docs lock and counters in the project data directory ([781bca1](https://github.com/EqualSoftware/claude-dx/commit/781bca11f7bde401a722fe7fc6b3453a4f66e887))
* **session:** hold session locks and remove unused plugin directories ([37a22c4](https://github.com/EqualSoftware/claude-dx/commit/37a22c4972b6bebb9916066ab01358b0d1682517))
* **spec:** add spec documents and the spec tools ([b787dce](https://github.com/EqualSoftware/claude-dx/commit/b787dce002be69c16e8c6ba2d861be190b378da5))
* **spec:** give the old slug of a split in spec_create ([cdb3800](https://github.com/EqualSoftware/claude-dx/commit/cdb3800b1a0bdf37638b51ef308094dac240b1be))
* **spec:** refuse spec_delete while the spec has local tasks ([3c13da4](https://github.com/EqualSoftware/claude-dx/commit/3c13da446311628e855a3a54589d81fc6138330a))
* **task:** add task milestones and the ready list ([396c715](https://github.com/EqualSoftware/claude-dx/commit/396c715f8a9a2d381d1bb1ebc079c462e5d33f76))
* **task:** add the task tools, INDEX.md and the /dx:task skill ([de61821](https://github.com/EqualSoftware/claude-dx/commit/de61821660fa9f7064db253bcfdfa82cac6c43c5))
* **task:** allow only the status changes of the task flow ([8a20853](https://github.com/EqualSoftware/claude-dx/commit/8a20853cc358dd890e56a78aa2bb4d156ecfa0a0))
* **task:** check task dependencies ([168f32f](https://github.com/EqualSoftware/claude-dx/commit/168f32f315e7af0da9e50f9aafb42d5285d5d850))
* **task:** move tasks to local files in the project data directory ([60a7ff1](https://github.com/EqualSoftware/claude-dx/commit/60a7ff1a69af2fa508008394a3ea06212ba1b3e9))
* **task:** protect controlled sections and add DoD tools ([d121c6e](https://github.com/EqualSoftware/claude-dx/commit/d121c6e7654c483da42c09d2565bbc45e127eb9d))
* **task:** review a task again after a scope change ([05c3179](https://github.com/EqualSoftware/claude-dx/commit/05c3179fa94a2e2d232c7776e5f38e0211eb3e8c))
* **task:** show the status of open dependencies in task_list ([67474cf](https://github.com/EqualSoftware/claude-dx/commit/67474cf6ccef3527876ef6ba3d021048623687a4))
* **task:** warn about open tasks on the same files ([36adf96](https://github.com/EqualSoftware/claude-dx/commit/36adf96827f82822cf720a47d247bd33bb254d84))


### Bug Fixes

* **cli:** keep the dev plugin of a running session ([e1366e3](https://github.com/EqualSoftware/claude-dx/commit/e1366e3e7dea753a816723daa43296b7f25b2141))
* **docs:** atomic writes and a docs lock for doc edits ([e60dcc5](https://github.com/EqualSoftware/claude-dx/commit/e60dcc5479d0ec67906ab7e3763416b6f22dee64))
* **docs:** escape pipes in README index rows ([169722e](https://github.com/EqualSoftware/claude-dx/commit/169722e2a972c27c8f10d3385d4fb17674fbf897))
* **docs:** keep omitted fields in glossary_upsert ([352d69f](https://github.com/EqualSoftware/claude-dx/commit/352d69fce1a3d13e0b7f330fe7245d6ebbca090c))
* **docs:** keep the file mode in docfile.Write ([870541c](https://github.com/EqualSoftware/claude-dx/commit/870541c88061f2c160848684ee0c44376f8c9b23))
* **docs:** lock the docs for read tools ([c33ca75](https://github.com/EqualSoftware/claude-dx/commit/c33ca7512c7f450e3ae14bb120d9bbf7654fac1c))
* **docs:** refuse a title with an empty file slug ([6c6a7d0](https://github.com/EqualSoftware/claude-dx/commit/6c6a7d0671c521ad87d76ff08b7633c08679dbcf))
* **docs:** stop ADR supersede cycles ([6f5017f](https://github.com/EqualSoftware/claude-dx/commit/6f5017fdcd8440c6f0f079a924688a1b7317a595))
* **docs:** warn about ADR and learning read errors ([fc8364a](https://github.com/EqualSoftware/claude-dx/commit/fc8364a0f6b4aee30820a24096bf3a8deae37437))
* **glossary:** refuse empty fields and bad Do not use words ([47f7b49](https://github.com/EqualSoftware/claude-dx/commit/47f7b490bc69d093346a5420b310a8d619a454ae))
* **install:** do not close the PowerShell window on failure ([3fe3655](https://github.com/EqualSoftware/claude-dx/commit/3fe36556fc3cba8789192e1442294a29de82a397))
* **markdown:** replace block values in SetField ([d2c11b8](https://github.com/EqualSoftware/claude-dx/commit/d2c11b80c5ede39cbedad86f9b1f2296bb6f8724))
* **mcp:** change only the title heading in task_update_fields ([37d7e7c](https://github.com/EqualSoftware/claude-dx/commit/37d7e7cd4e7f4b04c4c6a870a39e0edca3d127fd))
* **mcp:** use the git user.name as the default owner ([37ce562](https://github.com/EqualSoftware/claude-dx/commit/37ce562fdcd77ebef099ade1f37e64820d2926df))
* **plugin:** extract the plugin per binary path ([aef1390](https://github.com/EqualSoftware/claude-dx/commit/aef1390d1ee710143c8c9bf5896d57f875f66673))
* **plugin:** give full tool names in the dx:scout rules ([47e0d31](https://github.com/EqualSoftware/claude-dx/commit/47e0d31a4ca91e495b5cc4cce982b8e91193e8ef))
* **plugin:** let dx:scout give a dx tool as the source of a fact ([7063947](https://github.com/EqualSoftware/claude-dx/commit/7063947c66fc2571f31c1afe0d449355fb810f31))
* **plugin:** list tasks with task_list only in /dx:task ([40e462f](https://github.com/EqualSoftware/claude-dx/commit/40e462f07e0655616e8617fd69fb70a5818f12b9))
* **result:** align result tables by character ([472f06b](https://github.com/EqualSoftware/claude-dx/commit/472f06b3e689bf11e4076ccf8480c5fa65a28331))
* **session:** free session files of ended sessions ([4d238c1](https://github.com/EqualSoftware/claude-dx/commit/4d238c1d851477c59b4961cc1d23c2282c58736f))
* **session:** use a read-write cache lock ([2db7ef3](https://github.com/EqualSoftware/claude-dx/commit/2db7ef3a00aa8f26b4a93c7fe5a57db3991af446))
* **task:** check only the new task dependencies ([41093b4](https://github.com/EqualSoftware/claude-dx/commit/41093b4997066ba19d3639a7b468e13d2f75a78e))
* **task:** check section bodies of task tools ([2484d5e](https://github.com/EqualSoftware/claude-dx/commit/2484d5ea6a334785d2c271b92cd2a0c0158b80e5))
* **task:** keep one task ID when renumber fails ([06421cb](https://github.com/EqualSoftware/claude-dx/commit/06421cb6f7e89eae6227b9c221fcc4be6121d656))
* **task:** keep questions that start with a bracket ([098c18c](https://github.com/EqualSoftware/claude-dx/commit/098c18c3cbcc2d51fa9c5693c2d8e7aee150adce))
* **task:** refuse a used title in a title change ([4cff057](https://github.com/EqualSoftware/claude-dx/commit/4cff0575d5449ea3f45ea2435400de68d0c07bd3))
* **task:** remove the eureka item from the task template ([8a365b5](https://github.com/EqualSoftware/claude-dx/commit/8a365b55dcfc3bb7233bd663cde85edf4f541784))
* **task:** rename task files with one rename ([dfbfe34](https://github.com/EqualSoftware/claude-dx/commit/dfbfe346ccb9a419a28390f03635ec60c81574ce))
* **task:** rename the task file with the title ([6344108](https://github.com/EqualSoftware/claude-dx/commit/6344108f6248a2e3093fe0ab9ee1331d130a66f8))
