---
name: spike
description: Keep the question, findings and result of an exploration in a local spike with the dx spike tools. Use for an exploration before a spec exists, when a finding must not be lost, or when an exploration ends.
---

# Explore with a spike

A spike is a local record of one exploration: a Question, Notes and a Result. It is not in git. It has no tasks and no spec. The code of an exploration does not need an Approved spec. Do not change spike files with Edit or Write. Use the spike tools.

## When to create a spike

- Create a spike when the work is an exploration and no spec exists yet.
- Run `spike_list` first. If an Open spike has the same question, use it. Do not make a second one.
- Run `spike_create` with the slug (kebab-case, 40 characters or less), the title and the question. The question is one question that the exploration answers.
- Do not create a spike for work that an Approved spec or a task already covers.

## When to add a note

- Run `spike_note` with the slug and one finding when you learn a fact that answers part of the question.
- Write one finding on one line. The tool adds the date.
- Add the note when you learn the fact. Do not wait for the end.
- Do not note guesses or a history of your steps.

## When to close the spike

- Run `spike_close` with the slug and the result when the question has an answer, or when the exploration stops.
- The result answers the question. It says what you decided or what you could not find.
- A Closed spike is final. Do not reopen it. A new exploration is a new spike.

## Read and delete

- Use `spike_list` for the spikes, with the `status` filter if needed. Use `spike_get` to read one spike, or one section.
- Use `spike_delete` only when the user asks to delete a spike.

## From spike to spec

`/dx:spec` makes the spec from the spike and the code. The spike does not become a spec by itself. When the user wants a spec, tell the user to run `/dx:spec`, and read the spike with `spike_get` as the input.

## Report

The slug of each spike that you created, noted or closed, with one line of its result or its last note.
