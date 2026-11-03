package riidoaiserver

// promptHierarchyTerms maps the Riido screen names people use in requests to
// the hierarchy fields and CLI data. Users say 목표 or 상위 목표 for what the
// data calls milestone; without this the agent reports "no goal" while a
// milestone is present (RIID-6014 QA).
const promptHierarchyTerms = "Riido screen names: project = 프로젝트, milestone = 목표 (상위 목표), parent_task = 상위 작업, child tasks = 하위 작업. Read the user's Korean terms as these fields.\n"
