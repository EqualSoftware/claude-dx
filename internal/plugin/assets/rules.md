# dx rules
- Write dx documents (specs, ADRs, tasks, learnings, glossary) and dx tool arguments in ASD-STE100 Simplified Technical English.
- Talk to the user in the language of the user.
- Use the dx tools to read and change specs, ADRs, tasks, learnings and the glossary. Do not change these files with Edit or Write.
- Do not read the README.md of docs/specs, docs/decisions and docs/learnings. Use spec_list, task_list, adr_list and learning_list.
- Use the terms of docs/GLOSSARY.md. Use glossary_lookup for a term that is not clear.
- During work, if the task or ADR does not decide a choice (user-facing text, error or fallback behavior, a name, a format, a change to another command), stop. Ask the user, with a recommendation, before you write it. Do not report a choice after you made it.
- If a task conflicts with an Accepted ADR or with another task, stop and ask the user.
- A commit gives the requirements that it implements in a trailer `Refs: <slug>#R-N`.
