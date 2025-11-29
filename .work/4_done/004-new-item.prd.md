---
id: 004
title: New item flow
status: done
kind: prd
assigned: wkallan1984@gmail.com
estimate: 1 day
created: 2025-10-23T15:00:00Z
---
# New item flow with interactive input prompts enabled via --interactive flag

## Current experience
When I ran the following command:
```bash
$ kira new task todo "code-quality-controls"
```
I had the following experience - which:
```text
Enter step2 (Second step): foo
Enter done1 (First completion criterion): bar
Enter done2 (Second completion criterion): baz
Enter estimate (Estimate in days): 1
Enter assigned (Assigned to (email)): wkallan1984@gmail.com
Enter description (What needs to be done?): Set up code quality checks like linting, code, formating etc to ensure code stays consistent and readable.
Enter step1 (First step): ree
Enter step3 (Third step): ree
Enter release_notes (Public-facing changes (optional)):
Enter notes (Additional notes):
Enter tags (Tags):
1. implementation
2. maintenance
3. refactoring
Select option (number): 1
Created work item 003 in
```

Note - there was a bug in the code that caused the work item to be created in the wrong folder. e.g.The work item was then created at `.work/003-code-quality-controls.task.md` instead of `.work/1_todo/003-code-quality-controls.task.md` as expected.

## Desired experience

#### Scenario 1
When I run the following command without the --interactive flag:
```bash
$ kira new task todo "code-quality-controls"
```
Then I should get a file created under the todo folder with only the information from the initial command replaced in the template - the other template fields should be left as is.

#### Scenario 2
When I run the following command with the --interactive flag:
```bash
$ kira new task todo "code-quality-controls" --interactive
```
Then I should get a prompt for each template field that is not provided in the initial command.
