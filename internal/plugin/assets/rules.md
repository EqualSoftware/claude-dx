# dx rules
- Write dx documents (ADRs, tasks, learnings, glossary) and dx tool arguments in ASD-STE100 Simplified Technical English.
- Talk to the user in the language of the user.
- Use the dx tools to read and change ADRs, tasks, learnings and the glossary. Do not change these files with Edit or Write.
- Do not read the README.md of docs/tasks, docs/decisions and docs/learnings. Use task_list, adr_list and learning_list.
- Use the terms of docs/GLOSSARY.md. Use glossary_lookup for a term that is not clear.
- Give a fact search of more than 3 files to the dx:scout agent, with a closed list of questions.
- During work, if the task or ADR does not decide a choice (user-facing text, error or fallback behavior, a name, a format, a change to another command), stop. Ask the user, with a recommendation, before you write it. Do not report a choice after you made it.
- If a task conflicts with an Accepted ADR or with another task, stop and ask the user.
