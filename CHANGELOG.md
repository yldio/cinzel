cinzel
## [1.0.0](https://github.com/yldio/cinzel/compare/v0.7.0..v1.0.0) - 2026-09-25

### ⛰️  Features

- *(command)* [**breaking**] Exit non-zero when an action cannot be pinned or upgraded (#96) - ([53386ad](https://github.com/yldio/cinzel/commit/53386ad6188619b37ad459278cd810e5103a2286))

### 🐛 Bug Fixes

- *(ai)* Say when the config file exists and was not used - ([cb8aad6](https://github.com/yldio/cinzel/commit/cb8aad619033b076602c44a3ef1433995e351f0f))
- *(ai)* Drop a partial rune left by the context truncation - ([ca7b459](https://github.com/yldio/cinzel/commit/ca7b459d653d7e9308163dffa4f83680a4e3d198))
- *(ai)* Take only a fence at column 0 as a fence - ([1066023](https://github.com/yldio/cinzel/commit/10660234cc102307e99fe4558a006b6263e3a533))
- *(assist)* Give a clashing block a free label, and move its references with it (#83) - ([b6cbb8b](https://github.com/yldio/cinzel/commit/b6cbb8bf526c7809b81c7882a94e2f0cc257e2ef))
- *(cinzelerror)* Escape the bidi overrides that reorder an error message (#86) - ([a258f7f](https://github.com/yldio/cinzel/commit/a258f7ff2e2b8f5aabd64690029d6e430068eb3b))
- *(fsutil)* Read an HCL file whatever the case of its extension (#91) - ([8ce4c89](https://github.com/yldio/cinzel/commit/8ce4c898af6673a5e76589386e9a0254dbd3be89))
- *(fsutil)* Read a file it cannot open as one it does not own (#84) - ([0730f4d](https://github.com/yldio/cinzel/commit/0730f4d3736221600910314a9bb7de4a6ddcb120))
- *(github)* Keep the comment above a key that holds a map (#95) - ([82cc328](https://github.com/yldio/cinzel/commit/82cc32876efd940a27a4697dce07a1d643b32fe7))
- *(github)* Keep the comment above a jobs or steps list (#92) - ([28c41d0](https://github.com/yldio/cinzel/commit/28c41d0a0541f8bf6a963975832db85d7f8b17cc))
- *(github)* Read an env block's comments from both paths (#80) - ([a6d5e23](https://github.com/yldio/cinzel/commit/a6d5e2326ddddc85510aa713d69e06164df2eab5))
- *(github)* Keep a matrix axis named the way it was written - ([9f3d15b](https://github.com/yldio/cinzel/commit/9f3d15ba0f41e0cedd7c233350daedcd8f4aba7c))
- *(gitlab)* Read back the top-level key unparse passes through (#93) - ([c75c525](https://github.com/yldio/cinzel/commit/c75c525ce95a575623daf25bdb5dc52e5d50bc6d))
- *(gitlab)* Refuse the top-level key that sanitizes to itself (#88) - ([0d280f3](https://github.com/yldio/cinzel/commit/0d280f39d49cab483798130c371e3ef19e2c23f2))
- *(hclparser)* Keep the fraction when adding, subtracting or multiplying (#94) - ([4cfe6d8](https://github.com/yldio/cinzel/commit/4cfe6d83e54b8e56a5737b484e9692d5a4ad9d40))
- *(naming)* Prefix an identifier starting with any digit - ([bf413e0](https://github.com/yldio/cinzel/commit/bf413e00be4c502179e5349ea893bec932c98385))
- *(pin)* Count a file it could not read as a failure (#90) - ([1527df3](https://github.com/yldio/cinzel/commit/1527df337066b33bc9e6b72a4a5d835401aa7584))
- *(pin)* Recognise the tag a pass wrote by the shape of the comment (#82) - ([6ce3f1f](https://github.com/yldio/cinzel/commit/6ce3f1f8f90ffcf2f6353c51ff176c269036a1da))
- *(pin)* Read the whole trailing comment, and keep what the author wrote in it (#81) - ([98ab484](https://github.com/yldio/cinzel/commit/98ab4848bca5399e2eb834e3224a24ab96d8e5a3))
- *(unescape)* Keep a backslash the author wrote (#89) - ([7635482](https://github.com/yldio/cinzel/commit/76354825e74c6f441b8dc7d7cb0c0873379e85fd))
- *(yamldoc)* Quote a multi-line value a literal block cannot state (#87) - ([f4cedf3](https://github.com/yldio/cinzel/commit/f4cedf36d6303c3797f1c162b979564739fe2498))
- *(yamlwriter)* Take a cty.Value by whatever route it arrives - ([d1005b6](https://github.com/yldio/cinzel/commit/d1005b6a0aedf176c5c945f8a760afe2481fb531))

### 📚 Documentation

- Record the cty.Value-behind-an-any marshal bug - ([68e2aab](https://github.com/yldio/cinzel/commit/68e2aabe8d025f6a8937185163e3b1fd61555bf3))
- Record the five bugs from the internal/ai and naming review - ([788ca4f](https://github.com/yldio/cinzel/commit/788ca4f60be809434087dc058aed993675d77414))

### Cleanup

- *(yamlwriter)* Drop the unused Writer surface - ([1fbfd0a](https://github.com/yldio/cinzel/commit/1fbfd0afb4e1cf72a9ff04245ff8c9fc3df847ac))


## [0.7.0](https://github.com/yldio/cinzel/compare/v0.6.2..v0.7.0) - 2026-09-18

### ⛰️  Features

- *(github)* Carry comments into nested unparse bodies - ([39635ec](https://github.com/yldio/cinzel/commit/39635ec75ba730d5fbd4badeed8a850091f2be97))
- *(github)* Carry foot comments in both directions - ([62c2319](https://github.com/yldio/cinzel/commit/62c231901013a91134178472174acf886337065b))
- *(github)* Keep every comment written on a step - ([69a2963](https://github.com/yldio/cinzel/commit/69a29636d1b6141ec8e3ed59fe70f8a7a1ff4935))
- *(github)* Keep the comment written above any block - ([6332435](https://github.com/yldio/cinzel/commit/63324353898f18bcb874e8121a93f14f83b88edd))
- *(github)* Write attribute comments back to HCL - ([692504d](https://github.com/yldio/cinzel/commit/692504dade732d4408761104574aadd88693d302))
- *(github)* Keep the comment written above an attribute - ([091d411](https://github.com/yldio/cinzel/commit/091d41197a0d00e4c9e13698e6b00cc60ef006f2))
- *(github)* Keep the comment above a job block when parsing - ([7f1b114](https://github.com/yldio/cinzel/commit/7f1b1149fe96b6744f444f3d23b625c39cfa757a))
- *(github)* Keep the comment above a job when unparsing - ([9ab4c71](https://github.com/yldio/cinzel/commit/9ab4c71f495d2df1761b602127c9225ac4a1f2b6))
- *(gitlab)* Carry comments from YAML to HCL - ([55da191](https://github.com/yldio/cinzel/commit/55da19156c44ae88090b462d2ff1fbf76e907725))
- *(gitlab)* Carry comments from HCL to YAML - ([317067c](https://github.com/yldio/cinzel/commit/317067cea094007dfa015cc8938884a2773365c4))

### 🐛 Bug Fixes

- *(ai)* Say when a config holds an api_key, not only when it is exposed - ([470e2e3](https://github.com/yldio/cinzel/commit/470e2e390025b680c22e347ea2fdd7395fa5f5f2))
- *(assist)* Pick the last session by its timestamp, not its name length - ([52b78ba](https://github.com/yldio/cinzel/commit/52b78ba6901d5e03c25aea0472871db1ffffc8f0))
- *(assist)* Only treat a "---" at column 0 as a document separator - ([07acb5a](https://github.com/yldio/cinzel/commit/07acb5a8fad71a4f8ffe184dce2622d3ff58112f))
- *(github)* Report a NaN rather than panicking on one - ([8cff9f9](https://github.com/yldio/cinzel/commit/8cff9f9134b80d89603106e0d4a76a8f45e11c78))
- *(github)* Keep the comments written on an action's runs.env - ([9672180](https://github.com/yldio/cinzel/commit/96721807403bc0cab3bd7f4b60bd137ff88cf136))
- *(github)* Refuse two steps in a composite action sharing an id - ([17338f2](https://github.com/yldio/cinzel/commit/17338f241630206312a80935e122f207a8e2909b))
- *(github)* Read the step id through the comment wrapper - ([75a8773](https://github.com/yldio/cinzel/commit/75a87734ab4cbdb25b874dfbe2646363f4f6e54b))
- *(github)* Write a comment as it was written, not as we would write it - ([af9dd2b](https://github.com/yldio/cinzel/commit/af9dd2b0f123a5e17e01efe50164246c96686018))
- *(gitlab)* Refuse an artifacts or reports list of more than one - ([2c7d0bc](https://github.com/yldio/cinzel/commit/2c7d0bc8f1ea423f3e2b8fd8cf04e9c6cd61bdd3))
- *(gitlab)* Keep an explicit null services - ([fb1969b](https://github.com/yldio/cinzel/commit/fb1969bf3cf7946566baea9386a55852827eb650))
- *(gitlab)* Refuse a need block naming more than one job - ([1f3971e](https://github.com/yldio/cinzel/commit/1f3971e66e6b52a343ac605fd8bcdbcc7c74ec33))
- *(gitlab)* Apply the comment renames deterministically - ([4919517](https://github.com/yldio/cinzel/commit/4919517601d4dae2dc9e9da43a04a2ac5c7d2f56))
- *(gitlab)* Refuse a job named after a reserved keyword - ([e97dac4](https://github.com/yldio/cinzel/commit/e97dac41bba98f8dcfca1d07d5215bd3d0541702))
- *(hclparser)* Keep every part of a template - ([0331c55](https://github.com/yldio/cinzel/commit/0331c55df3f830e320a13f40ccde472155cdc5a9))
- *(hclparser)* Refuse a non-number operand instead of panicking - ([e08db52](https://github.com/yldio/cinzel/commit/e08db52da19b0a42e389fe4c2629f04d5db2c892))
- *(hclparser)* Read a comment HCL wrote with any of its markers - ([08a2792](https://github.com/yldio/cinzel/commit/08a27924e0043d85cdfd1f597d6789e73f6c82d3))
- *(pin)* Report what cannot be pinned instead of calling it pinned - ([656536d](https://github.com/yldio/cinzel/commit/656536d635fcc79b1e676719335d454824b11112))
- Refuse configuration paths that name one machine - ([001ecb0](https://github.com/yldio/cinzel/commit/001ecb012e42ae46cc32442ff8cd5801f5799e11))
- Stop telling authors to file a bug over their own typos - ([1d425c8](https://github.com/yldio/cinzel/commit/1d425c8bb812e37c567b1de99c5adc2433c8359d))

### 🚜 Refactor

- Move the comment collector to where every parse can reach it - ([77659b9](https://github.com/yldio/cinzel/commit/77659b9fc31d97c8fc2d3cd9a503352e0a23e081))
- Read inline comments from what was parsed, not from disk - ([e4d9dec](https://github.com/yldio/cinzel/commit/e4d9dec16a2f9ad3502166f4adfd5816c46995de))

### 📚 Documentation

- Record the four bugs worth a solution note - ([3fd2ab0](https://github.com/yldio/cinzel/commit/3fd2ab031c588249ffcb970995dfe4b1be8fde3c))
- Bring the README back in step with the commands - ([2b9fd2a](https://github.com/yldio/cinzel/commit/2b9fd2a8f3c542995a7ca990829ef84c005da260))
- Move the reference detail out of CLAUDE.md into docs/architecture - ([a5b66c4](https://github.com/yldio/cinzel/commit/a5b66c4fafdba807e49cdd51f907f5468ea80008))
- Rewrite CLAUDE.md around what the code actually does - ([f472603](https://github.com/yldio/cinzel/commit/f47260319c225e59a3378c32a48669aaad80479f))
- Drop winget and stop offering api_key as an option - ([60de39f](https://github.com/yldio/cinzel/commit/60de39f5c07980547dd51a5bf777343fa9e8dc2c))
- Reconcile the solution notes with the code they describe - ([3f692a0](https://github.com/yldio/cinzel/commit/3f692a0757588ea252a72b177a242f8fcf0f4343))
- Say where the deleted plans went - ([2495008](https://github.com/yldio/cinzel/commit/249500800ff4dd970ba1f7e3966bab4577be6b2b))
- Remove the brainstorms and the release follow-up plan - ([1a755dd](https://github.com/yldio/cinzel/commit/1a755dd7a6db13e73ab7ba679a69f84b993c27ad))
- Remove the assist plan - ([64f8348](https://github.com/yldio/cinzel/commit/64f8348b41fe494678a809151446f615117a7906))
- Remove the comment fidelity plan - ([088565f](https://github.com/yldio/cinzel/commit/088565f6b475953fc601537b6a6c5ae4d9a32a54))
- Remove the completed plans - ([5a115de](https://github.com/yldio/cinzel/commit/5a115def4a15a4e476e2b531bbb2fe59759fbf41))
- Record the GitHub nested-body commit in the plan - ([1b5e55e](https://github.com/yldio/cinzel/commit/1b5e55ef88a9b5e131e77c9d6850a3117a2b878a))
- Record the GitLab commits and split item 8 in the plan - ([18c3ea6](https://github.com/yldio/cinzel/commit/18c3ea6ac8fd7ad871a87160436616e9eac66730))
- Record the marker commit and the foot comment one in the plan - ([42cf210](https://github.com/yldio/cinzel/commit/42cf21095fbf78f8d5a34f576235bf62fe7a6f86))
- Plan comment fidelity across both providers - ([32a4a36](https://github.com/yldio/cinzel/commit/32a4a3688c16f1d98cdce8d967b321228092f3dc))

### 🧪 Testing

- *(assist)* Cover the session picked by --refine - ([5fa62e3](https://github.com/yldio/cinzel/commit/5fa62e3c5a25899ce3fb68cc5efe2d64cc72e757))
- Make the main subtest pass on its own - ([4b65665](https://github.com/yldio/cinzel/commit/4b65665795e7a32cbcbd0a2bf25851d82ae08efa))
- Close the roundtrip gaps in the comment acceptance list - ([eaed2d5](https://github.com/yldio/cinzel/commit/eaed2d5c339e5636905d215bdb7ca67711300c5f))


## [0.6.2](https://github.com/yldio/cinzel/compare/v0.6.1..v0.6.2) - 2026-09-16

### 🐛 Bug Fixes

- *(github)* Say where two step blocks share a label - ([68766ae](https://github.com/yldio/cinzel/commit/68766aefce2e9b8f2133e59b2c1a7d5322650b66))


## [0.6.1](https://github.com/yldio/cinzel/compare/v0.6.0..v0.6.1) - 2026-09-16

### 🐛 Bug Fixes

- *(github)* Keep unparse output readable by parse - ([63491ba](https://github.com/yldio/cinzel/commit/63491ba2252f4e1acf6720f7f68c0f5d95485081))
- *(github)* Refuse two steps in one job writing the same step id - ([2b02924](https://github.com/yldio/cinzel/commit/2b02924a2d158cc8d07befa75482355f3d27b35a))

### 📚 Documentation

- Record the deferred embedded actionlint investigation - ([d1a7ce7](https://github.com/yldio/cinzel/commit/d1a7ce7bd4d239ea0955a2523f66f39f960b3386))


## [0.6.0](https://github.com/yldio/cinzel/compare/v0.5.6..v0.6.0) - 2026-09-16

### ⛰️  Features

- *(ai)* Warn about a config others can read (#79) - ([4ca7309](https://github.com/yldio/cinzel/commit/4ca73096792b5c372893e2e788ebef031d256f81))

### 🐛 Bug Fixes

- *(ai)* Take the default models from the SDKs - ([23d7bf6](https://github.com/yldio/cinzel/commit/23d7bf6ec60c78a6990aba2d3f41a6dff45f80bd))
- *(ai)* Classify provider errors from the status code - ([8b6c270](https://github.com/yldio/cinzel/commit/8b6c2707fc95ab8ee71410fc2e4fcb147bd3f99b))
- *(assist)* Only treat a leading ".." element as traversal - ([1e93f89](https://github.com/yldio/cinzel/commit/1e93f89632a71bbffac1e3a9496923ba4c195ab6))
- *(assist)* Skip comments when reading a block signature - ([d9ca538](https://github.com/yldio/cinzel/commit/d9ca53878dbbc016f64620453a6e5b455bfd5440))
- *(assist)* Read the API key env var first (#74) - ([4e6cd3b](https://github.com/yldio/cinzel/commit/4e6cd3bbf48c92d77e35327af542aac152d28ee4))
- *(cli)* Escape control characters before an error reaches the terminal (#57) - ([2fc5fe5](https://github.com/yldio/cinzel/commit/2fc5fe5b8e861e82c707b563cd50143bddcf2890))
- *(errors)* Fall back to a diagnostic's summary - ([6e4f220](https://github.com/yldio/cinzel/commit/6e4f220acfc3ca6e5b8600eb2b006ebe1dc87b7b))
- *(github)* Detect an action without requiring a name - ([2cafc50](https://github.com/yldio/cinzel/commit/2cafc5042104666586ad4a15b0613849db6b6dea))
- *(github)* Stop reading JSON braces as an orphaned closer - ([1a738c9](https://github.com/yldio/cinzel/commit/1a738c95bcb822c090fbb753fb6e2360e9963181))
- *(github)* Check cron on the parse path, and accept names - ([c72427e](https://github.com/yldio/cinzel/commit/c72427e04de895270d43cccb6245cda1a75e9ca5))
- *(github)* Validate a workflow once its jobs are in it - ([2ac6579](https://github.com/yldio/cinzel/commit/2ac65796afe395ab1cf06b19651772b58f83267e))
- *(github)* Keep an authored id in the step fingerprint - ([8e5a825](https://github.com/yldio/cinzel/commit/8e5a825e266dfd9100f1ec1f5df197eea8a9e672))
- *(github)* Spell deprecationMessage the way GitHub does - ([5dba9a3](https://github.com/yldio/cinzel/commit/5dba9a3a4fb498d385bee191f5bc040e834cab33))
- *(github)* Stop handing two steps the same label - ([c3f3627](https://github.com/yldio/cinzel/commit/c3f362765cf8869b01dc7c1adb81c228dd76323a))
- *(github)* Reject two jobs that write the same key - ([55d8540](https://github.com/yldio/cinzel/commit/55d8540a2266508a4cff74237d8259617964b2ac))
- *(github)* Mark an emitted action so a renamed one is pruned (#68) - ([445aaf4](https://github.com/yldio/cinzel/commit/445aaf4391d29ee42e72d47265adcb14dc222462))
- *(github)* Keep an action's name when unparsing it (#66) - ([2477964](https://github.com/yldio/cinzel/commit/2477964b84df5d6195b2c7dfa441fddd78077eb2))
- *(github)* Prune a stale generated file in a subdirectory (#65) - ([ce8c0f9](https://github.com/yldio/cinzel/commit/ce8c0f98e7b5a1dee606d260a9cec532a2f64b22))
- *(github)* Refuse two definitions writing to the same file (#64) - ([1066b8f](https://github.com/yldio/cinzel/commit/1066b8f765b25dcb325ce621ae0a429f704cc0ad))
- *(github)* Keep a filename inside the output directory (#63) - ([a786403](https://github.com/yldio/cinzel/commit/a78640324cc3f448b8873b897f64d4fe074467cf))
- *(gitlab)* Drop the empty root document from a spec-only pipeline - ([92feabe](https://github.com/yldio/cinzel/commit/92feabe7e2d50e47e7e8a27a32ab3cc4a824c618))
- *(gitlab)* Refuse a passthrough key that is not an identifier - ([6f4b068](https://github.com/yldio/cinzel/commit/6f4b068071f2aed3c12cc7acb8222ad09272b014))
- *(gitlab)* Declare parallel on a need block - ([7c0f8b7](https://github.com/yldio/cinzel/commit/7c0f8b72a4d3f5de97507b3bcff203dfd137c90c))
- *(gitlab)* Report a dropped workflow key, check top-level services - ([02f1900](https://github.com/yldio/cinzel/commit/02f1900c549fd4b71fbff4db713bbf4637707529))
- *(gitlab)* Refuse an artifacts list at write time - ([3580031](https://github.com/yldio/cinzel/commit/3580031ac6e3b820869f42deaffe9fdeb3eb5526))
- *(gitlab)* Remap the job name inside an object need - ([da84f6c](https://github.com/yldio/cinzel/commit/da84f6c0639099639d95f1b92ffb844e4dd14180))
- *(gitlab)* Reject a job named after a pipeline keyword - ([da64279](https://github.com/yldio/cinzel/commit/da6427962538e7bfc4e610a1bf9a52d15e3f3320))
- *(gitlab)* Stop collapsing an empty mapping into a null - ([23622aa](https://github.com/yldio/cinzel/commit/23622aadf3587c848b86d889f9f2833f59eec593))
- *(gitlab)* Refuse a template key with nothing after the dot (#73) - ([c971e9c](https://github.com/yldio/cinzel/commit/c971e9c8298522043fc2397396d6c598b7719ff8))
- *(gitlab)* Refuse a needs entry that names no job (#72) - ([ee5c11d](https://github.com/yldio/cinzel/commit/ee5c11dbdd6c5a5a867e72c6de67631ee8da7985))
- *(gitlab)* Refuse an unparse that converted nothing (#67) - ([b236f63](https://github.com/yldio/cinzel/commit/b236f638af836b85bcce2c3130b536d6fd8c1316))
- *(gitlab)* Refuse input that is not valid UTF-8 (#62) - ([2c48b19](https://github.com/yldio/cinzel/commit/2c48b19da13d43740384fb7927be9eaca6c900d8))
- *(gitlab)* Reject a key that is not a string, and a job with no name (#61) - ([1e25e11](https://github.com/yldio/cinzel/commit/1e25e1113371cbbf8a22b504ee0e87647d19c883))
- *(gitlab)* Refuse YAML whose aliases expand without bound (#60) - ([28525cf](https://github.com/yldio/cinzel/commit/28525cfa26a3f1f5e791b4cbe6bd75f32af1da9b))
- *(gitlab)* Keep a spec header, a run job and the global default keywords (#52) - ([49234a4](https://github.com/yldio/cinzel/commit/49234a4c240d8e1d74471ccc3e350814f35d6207))
- *(gitlab)* Keep a null rules, artifacts, only or except (#50) - ([3a15b2c](https://github.com/yldio/cinzel/commit/3a15b2cfa746dabdfe154780950bd4ad651f05cd))
- *(gitlab)* Keep an explicitly empty needs, cache, services or rules (#48) - ([b6c2520](https://github.com/yldio/cinzel/commit/b6c2520e3bf4d2d766c046a69c74280d0a2dd921))
- *(gitlab)* Let a workflow rule carry its own auto_cancel (#41) - ([cc8af29](https://github.com/yldio/cinzel/commit/cc8af29e7921444b0915d4b1c8a12c0e1bf3a570))
- *(gitlab)* Declare a job's inputs and publish (#45) - ([ce7249a](https://github.com/yldio/cinzel/commit/ce7249a751a49279de6e9bea092acad72c9ad1b2))
- *(gitlab)* Accept a script given as a single string (#43) - ([0806c32](https://github.com/yldio/cinzel/commit/0806c322c972c3bcf3518b25b0bb52271854e4d8))
- *(gitlab)* Carry only/except and a variable's expand through (#39) - ([acbef33](https://github.com/yldio/cinzel/commit/acbef3346e09b791c461a8b9be9ac108abbffad6))
- *(gitlab)* Treat a pipeline of only includes as a pipeline (#37) - ([f491753](https://github.com/yldio/cinzel/commit/f491753ce8a617644f039a1bb88df44ae4e3ca72))
- *(gitlab)* Declare the documented GitLab keywords in the HCL schema (#34) - ([7ed949b](https://github.com/yldio/cinzel/commit/7ed949bd09f6af5364832412767ea10a2f5aa870))
- *(gitlab)* Support several caches on a job or default (#28) - ([3e1cc09](https://github.com/yldio/cinzel/commit/3e1cc092f068a4a18147ae66f89650ae1a4554ca))
- *(gitlab)* Accept 'variables' and 'needs' in rule blocks (#29) - ([978777a](https://github.com/yldio/cinzel/commit/978777ac7fd564adabbcb0d76b1a916926992bbf))
- *(gitlab)* Support 'needs' entries in object form (#30) - ([21422ea](https://github.com/yldio/cinzel/commit/21422ea4aa2fd6de3656c05110cec145e520403b))
- *(gitlab)* Write nested maps as blocks only where the schema declares one (#26) - ([4419da4](https://github.com/yldio/cinzel/commit/4419da4be0a94fce09fb6f1617fd42c34c5bcfab))
- *(gitlab)* Treat a job without its own script as a job (#27) - ([e20cbfb](https://github.com/yldio/cinzel/commit/e20cbfbd9ef15d37743ef38f7456fe756238e04c))
- *(gitlab)* Write a leading "@" with double quotes (#25) - ([a7dbae4](https://github.com/yldio/cinzel/commit/a7dbae461ac20e4f1c74342601200371010fa701))
- *(parse)* Reject an input that declares nothing (#53) - ([8d72fad](https://github.com/yldio/cinzel/commit/8d72fadf7914446c312e74cdd2b0505c90afdbda))
- *(pin)* Refuse a resolve that is not a commit SHA - ([ba58454](https://github.com/yldio/cinzel/commit/ba58454e5f59025018089256d0cfb167380a3f25))
- *(pin)* Read uses blocks from the parse, not from the text (#71) - ([dcad045](https://github.com/yldio/cinzel/commit/dcad04537f041fd447e83bbc3d0612bf2489246b))
- *(pin)* Write a resolved SHA to the action it belongs to (#69) - ([48ed324](https://github.com/yldio/cinzel/commit/48ed324266f9a7191ccf9eb3ca8581eba4d6e83b))
- *(step)* Pick a heredoc marker the script does not use - ([20c3523](https://github.com/yldio/cinzel/commit/20c3523f827d8919bea8de3f23d93bde87f7f161))
- *(unicode)* Stop rewriting an escape the author wrote - ([2f65651](https://github.com/yldio/cinzel/commit/2f65651fff60384e05cd5091e95684c0f261a34f))
- *(unparse)* Stop two inputs writing to one output file - ([a9e18bb](https://github.com/yldio/cinzel/commit/a9e18bbaeb48ae903a4090210912530b8f874315))
- *(unparse)* Reject a key that is not a string, and resolve one that is an alias (#59) - ([b6221f1](https://github.com/yldio/cinzel/commit/b6221f1d8c6278ec2a0f32f71266d19f072637ca))
- *(unparse)* Keep a whole number too large for an integer exact (#58) - ([9ed4be6](https://github.com/yldio/cinzel/commit/9ed4be6cc284290d630205e1b45706f82f700c43))
- *(yaml)* Close three gaps in the quoting rule - ([ee7454d](https://github.com/yldio/cinzel/commit/ee7454d32fda20f497e8d883f4b76003241697b1))
- Fifteen conversion, CLI and repo defects (#54) - ([22c4de3](https://github.com/yldio/cinzel/commit/22c4de3e42f070cb265463ee433af9ed9c73f8a7))

### 📚 Documentation

- Record the workflow key order and the parse defaults - ([098724d](https://github.com/yldio/cinzel/commit/098724d32e1059ff860913252fec4d6d2dca52a6))

### ⚡ Performance

- *(unparse)* Build job identifiers against a set, not a slice (#56) - ([7474995](https://github.com/yldio/cinzel/commit/7474995b82e11321ae39dcf765557a07fcfc64c2))

### 🧪 Testing

- *(pin)* Stop the upgrade tests calling the live GitHub API (#70) - ([239a423](https://github.com/yldio/cinzel/commit/239a423db774b8416d682949d74d9f83f08f3be1))

### ⚙️ Miscellaneous Tasks

- *(config)* Drop two keys that did nothing (#78) - ([0c3079c](https://github.com/yldio/cinzel/commit/0c3079c2e192bd6b523e2383c067b4595b10a03c))
- *(init)* Stop writing API keys to the config (#77) - ([aca3b25](https://github.com/yldio/cinzel/commit/aca3b254246572b812f4f85b0c4fe0d1a03b075f))
- Check release token scope against the installation - ([4982199](https://github.com/yldio/cinzel/commit/49821990b5961b0b67f54428b2eacfeacd0021ca))
- Delete filereader and filewriter - ([60746c5](https://github.com/yldio/cinzel/commit/60746c50e05f2692aacbb1c38d30083495e9b1fc))
- Let the licence check skip the cask (#76) - ([da69af6](https://github.com/yldio/cinzel/commit/da69af6c2d65a6888b0e7ec57221f4936527166e))
- Add the missing licence headers (#75) - ([a1c62cb](https://github.com/yldio/cinzel/commit/a1c62cbcc53327de3bdefa6b2b514c9e8f969378))
- Resolve the ghalint findings and gate on them (#55) - ([3a8ee45](https://github.com/yldio/cinzel/commit/3a8ee459f7e6271d21511ee20dfbffae608dd5f0))
- Move the toolchain to Go 1.27 and git-cliff 2.14.1 (#46) - ([c02127d](https://github.com/yldio/cinzel/commit/c02127dfb02475c64bb27b56e06cc98a51c05e70))
- Release only on manual dispatch (#22) - ([54f2174](https://github.com/yldio/cinzel/commit/54f217428231ec90acd7907c3b9e40891f69f407))


## [0.5.6](https://github.com/yldio/cinzel/compare/v0.5.5..v0.5.6) - 2026-09-11

### 🐛 Bug Fixes

- *(gitlab)* Keep job and template names that are not HCL identifiers (#21) - ([6a681fc](https://github.com/yldio/cinzel/commit/6a681fc3d36e4fdd12e081d0d75a468b5628c1f3))


## [0.5.5](https://github.com/yldio/cinzel/compare/v0.5.4..v0.5.5) - 2026-09-11

### 🐛 Bug Fixes

- *(yamldoc)* Write a leading "@" with double quotes (#19) - ([a6d29da](https://github.com/yldio/cinzel/commit/a6d29da132adcb60a4ca110d79edbfe89296fab9))


## [0.5.4](https://github.com/yldio/cinzel/compare/v0.5.3..v0.5.4) - 2026-09-11

### 🐛 Bug Fixes

- *(github)* Reject nested block keys that cannot be written back (#18) - ([00c4e00](https://github.com/yldio/cinzel/commit/00c4e00d043129e1e5f8698adb9107c1a72a19e6))


## [0.5.3](https://github.com/yldio/cinzel/compare/v0.5.2..v0.5.3) - 2026-09-11

### 📚 Documentation

- *(github)* Correct why job keys are renamed (#16) - ([29c259e](https://github.com/yldio/cinzel/commit/29c259ec14308a90b8ea43dac74565bf8885d966))


## [0.5.2](https://github.com/yldio/cinzel/compare/v0.5.1..v0.5.2) - 2026-09-11

### 🐛 Bug Fixes

- *(github)* Keep job keys that are not valid HCL labels (#15) - ([624e214](https://github.com/yldio/cinzel/commit/624e214f6629d66cbcdb1050073183d36f6a66b5))


## [0.5.1](https://github.com/yldio/cinzel/compare/v0.5.0..v0.5.1) - 2026-09-11

### 🚜 Refactor

- *(github)* Emit workflow YAML from an ordered document, and fix three data-loss bugs (#12) - ([aff44cd](https://github.com/yldio/cinzel/commit/aff44cd65f8790b0189388c5cb5ba01f408b1caa))

### 📚 Documentation

- Add cinzelrc path portability across OS - ([0f5b9fb](https://github.com/yldio/cinzel/commit/0f5b9fb657053b7bdd70096be8c273e1534f86c2))


## [0.5.0](https://github.com/yldio/cinzel/compare/v0.4.0..v0.5.0) - 2026-04-13

### ⛰️  Features

- Propagate step uses version inline comment to YAML output (#11) - ([88fa46c](https://github.com/yldio/cinzel/commit/88fa46c302df7db6e9302b48eabea068234b5904))


## [0.4.0](https://github.com/yldio/cinzel/compare/v0.3.5..v0.4.0) - 2026-04-10

### ⛰️  Features

- Propagate HCL inline comments on attributes to YAML output (#10) - ([90ea5d5](https://github.com/yldio/cinzel/commit/90ea5d533954f6364ab108b4245c150f1e36fc25))


## [0.3.5](https://github.com/yldio/cinzel/compare/v0.3.4..v0.3.5) - 2026-04-10

### 🐛 Bug Fixes

- *(parse)* Always emit permissions: {} in workflow YAML output (#9) - ([fc0b359](https://github.com/yldio/cinzel/commit/fc0b35912bb0ed8ff133e1a9c5692e47f8950b1a))


## [0.3.4](https://github.com/yldio/cinzel/compare/v0.3.3..v0.3.4) - 2026-04-10

### 🐛 Bug Fixes

- Version comment and permissions (#8) - ([41c8925](https://github.com/yldio/cinzel/commit/41c89255811cb2cd4d6843bdbc3735e4e553626f))


## [0.3.3](https://github.com/yldio/cinzel/compare/v0.3.2..v0.3.3) - 2026-04-01

### 🐛 Bug Fixes

- Preserve HCL-defined job order in parse direction YAML output (#7) - ([8d27137](https://github.com/yldio/cinzel/commit/8d27137ca7a11619116860ab420507b402f136ae))


## [0.3.2](https://github.com/yldio/cinzel/compare/v0.3.1..v0.3.2) - 2026-04-01

### 🐛 Bug Fixes

- Preserve job order during YAML→HCL unparse using yaml.v3 Node API  (#6) - ([13cac9b](https://github.com/yldio/cinzel/commit/13cac9b2aa5533ee4a12757179b589abfb9ef134))


## [0.3.1](https://github.com/yldio/cinzel/compare/v0.3.0..v0.3.1) - 2026-03-31

### 🐛 Bug Fixes

- Preserve emoji ZWJ sequences through YAML↔HCL roundtrip (#5) - ([d4925da](https://github.com/yldio/cinzel/commit/d4925dacdb1be1ff1681741950b125a031fa76b5))


## [0.3.0](https://github.com/yldio/cinzel/compare/v0.2.0..v0.3.0) - 2026-03-31

### ⛰️  Features

- Add --yml flag to generate .yml workflow files instead of .yaml (#4) - ([a4b52ec](https://github.com/yldio/cinzel/commit/a4b52ecbeef0e05f7c40fea2fc1dd11bc42bc6e3))


## [0.2.0](https://github.com/yldio/cinzel/compare/v0.1.0..v0.2.0) - 2026-03-18

### ⛰️  Features

- AI assist command and github pin and upgrade subcommands (#3) - ([1539232](https://github.com/yldio/cinzel/commit/1539232429e89c8acfdb41183c0bba029d6f0ffe))


## [0.1.0](https://github.com/yldio/cinzel/compare/v0.0.8..v0.1.0) - 2026-03-13

### ⛰️  Features

- Auto-release on push to main - ([71b2105](https://github.com/yldio/cinzel/commit/71b210552dcf5b04ec289193dc4b0a17f5ab5bf5))
- Auto-release on push to main via semver - ([90f84e0](https://github.com/yldio/cinzel/commit/90f84e003feb508a26f8a7904a0195e6bd5bf1e7))

### 🐛 Bug Fixes

- Strip quarantine flag on macOS cask install - ([29f5708](https://github.com/yldio/cinzel/commit/29f5708b85f655e3ffc73170fe48b488696f8402))


## [0.0.8](https://github.com/yldio/cinzel/compare/v0.0.7..v0.0.8) - 2026-03-13

### ⛰️  Features

- Add Claude Code plugin with parse/unparse skills - ([42d6a2c](https://github.com/yldio/cinzel/commit/42d6a2ce2886db45799b4d6023dd5d4489a93445))

### 🐛 Bug Fixes

- Use --unreleased for release notes body - ([0a52381](https://github.com/yldio/cinzel/commit/0a523819e29877e5dba14c87963869bfbfb9fc12))


## [0.0.7](https://github.com/yldio/cinzel/compare/v0.0.6..v0.0.7) - 2026-03-12

### 🐛 Bug Fixes

- Scope app token to include homebrew tap - ([3619d49](https://github.com/yldio/cinzel/commit/3619d4939341d14aef6143b2d7c56f762b61d116))


## [0.0.6](https://github.com/yldio/cinzel/compare/v0.0.5..v0.0.6) - 2026-03-12

### 🐛 Bug Fixes

- Use --latest for release notes body - ([c3953b3](https://github.com/yldio/cinzel/commit/c3953b385e0a065f30e3c5b268e6db60520c3aa2))


## [0.0.5](https://github.com/yldio/cinzel/compare/v0.0.4..v0.0.5) - 2026-03-12

### 🐛 Bug Fixes

- Use homebrew-cinzel tap repo name - ([a0d4c0c](https://github.com/yldio/cinzel/commit/a0d4c0ca2324bd7b4fb8a8fade0f336da86aba34))


## [0.0.4](https://github.com/yldio/cinzel/compare/v0.0.3..v0.0.4) - 2026-03-12

### 📚 Documentation

- Capture release auth contract fix - ([d187c10](https://github.com/yldio/cinzel/commit/d187c1003c03ccdd84ba290e55e6f4b3cd054bfc))


## [0.0.3](https://github.com/yldio/cinzel/compare/v0.0.2..v0.0.3) - 2026-03-12

### 🐛 Bug Fixes

- Map app token for goreleaser homebrew - ([2bfc2b8](https://github.com/yldio/cinzel/commit/2bfc2b83cf88d7790ec551a6193a54865c896982))


## [0.0.2](https://github.com/yldio/cinzel/compare/v0.0.1..v0.0.2) - 2026-03-12

### 🐛 Bug Fixes

- Pass github token to tag action - ([2299c71](https://github.com/yldio/cinzel/commit/2299c71c10de0bdd46fadaafe4748fb78e0661e0))
- Restore github-tag-action token input - ([74b80ff](https://github.com/yldio/cinzel/commit/74b80ffab271ab115f406ee2ac8e1a9a1246e851))
- Use app token inputs for release actions - ([08bc6f3](https://github.com/yldio/cinzel/commit/08bc6f30e4ffb292a7652d0a43271f32f6c4a50a))


## [0.0.1] - 2026-03-12

### ⛰️  Features

- Harden release workflow and docs - ([36b136a](https://github.com/yldio/cinzel/commit/36b136a567a3fb22b946d1c605348ddb7a1a79be))
- Switch release workflow to GoReleaser - ([5a8bbca](https://github.com/yldio/cinzel/commit/5a8bbcac4c493fd64834a4c48d958ea83fe08f4d))
- Type gitlab parse schema and services - ([f5b32c4](https://github.com/yldio/cinzel/commit/f5b32c4c27c898fe02867f03d9c92f9e4a98f523))
- Support gitlab template extends and include - ([b76fdbf](https://github.com/yldio/cinzel/commit/b76fdbf4586d8604f7b8b238f98b12d3520f74cb))
- Add gitlab provider parse and unparse baseline - ([82deed2](https://github.com/yldio/cinzel/commit/82deed27ea218d03441f69bce23e6c11e7f506a3))
- Enforce strict github block schemas - ([8f79a6b](https://github.com/yldio/cinzel/commit/8f79a6bd75e40ede42dbca92464c99819d0d1961))
- Add cinzelrc command config precedence - ([4c46282](https://github.com/yldio/cinzel/commit/4c4628227ce7d762df6be779bf4a8a31a379775c))
- Refactor yamlparser as generic - ([6688bc2](https://github.com/yldio/cinzel/commit/6688bc2900916b4e89cacf17bd6ca3ffee48c89e))
- Refactor filereader as generic - ([27c9197](https://github.com/yldio/cinzel/commit/27c9197b1a17f462978b50611ba4208fb72cc540))
- Refactor to introduce providers - ([a60fdb4](https://github.com/yldio/cinzel/commit/a60fdb4790d197a856901534bf852f2a7497feec))
- Expand release automation flow - ([a431d06](https://github.com/yldio/cinzel/commit/a431d06232619d6c5c3a09e19c72c8a5602ebcf8))
- Add release automation scaffolding - ([2771259](https://github.com/yldio/cinzel/commit/27712597af1eaa3f1aa81711ce8f854f19c9d20f))
- Yaml to hcl - ([4920f3e](https://github.com/yldio/cinzel/commit/4920f3e1a990dab258f32c841e36c29cd974aa9d))
- Yaml to hcl - ([ceb491a](https://github.com/yldio/cinzel/commit/ceb491a3ff1e77375b2ff8889f8c9544976bd952))
- Yaml to hcl - ([9c8844b](https://github.com/yldio/cinzel/commit/9c8844b26c0b97cf31bac72b9019a639ff95820a))
- Yaml to hcl - ([a768192](https://github.com/yldio/cinzel/commit/a76819288ccf8240593470b99c4c32d2c9fee7dd))
- Yaml to hcl - ([652a67d](https://github.com/yldio/cinzel/commit/652a67d86903a47e294ea5535697f25469186bc3))
- Templates - ([59ed63c](https://github.com/yldio/cinzel/commit/59ed63c60f5ec070b8ff97100cc676000388e423))
- Templates - ([648340d](https://github.com/yldio/cinzel/commit/648340d57b5f79c384fec64ec4cba8a7ed68be29))
- Templates - ([f90d492](https://github.com/yldio/cinzel/commit/f90d4925c0139112df43a06a61dbf55f1d5d7015))
- Possibility to use variables (#26) - ([f34290b](https://github.com/yldio/cinzel/commit/f34290b79670f4470f8c16a0727e1fa7ebbd7a63))
- Implement missing workflow features - ([ad8a877](https://github.com/yldio/cinzel/commit/ad8a877eb93db3435dca63d23f721d1d5ac53631))
- Add-cache-to-go-setup (#6) - ([bfe83e2](https://github.com/yldio/cinzel/commit/bfe83e2bf2636fda97014a21a9f92f890bcf5839))
- Add Id for job resource (#3) - ([e64e3c4](https://github.com/yldio/cinzel/commit/e64e3c4e035818a5e731cecc81db988e66c2794e))
- Initial commit - ([66a8b52](https://github.com/yldio/cinzel/commit/66a8b529a57e102663c535cc1512c8ba51d63944))

### 🐛 Bug Fixes

- Normalize manual release version input - ([1d80bde](https://github.com/yldio/cinzel/commit/1d80bde61cfc4b8cfbd2b4b41713a4c394ac5630))
- Stabilize release workflow automation - ([8a1c1d1](https://github.com/yldio/cinzel/commit/8a1c1d1b25ab5d26216eb5d3c3455537df588843))
- Require explicit manual release tag - ([7af24c1](https://github.com/yldio/cinzel/commit/7af24c1233232365eba6a4b3eff67e47a4bd5119))
- Allow changelog push in manual release - ([1624886](https://github.com/yldio/cinzel/commit/1624886a666c89ce1ea78589114d603b05ba5605))
- Harden git-cliff release changelog step - ([3444209](https://github.com/yldio/cinzel/commit/3444209e11f91ccac26812f5b2ba33b168480009))
- Centralize generated marker ownership checks - ([aab0c88](https://github.com/yldio/cinzel/commit/aab0c88c0e773e9e0d66fd0c2150af081458adfa))
- Point cask repo to cinzel - ([9c54d7f](https://github.com/yldio/cinzel/commit/9c54d7fda844b18ce90d0f6e7b18b3dad60fc1e2))
- Align versioning across local and release builds - ([5f9c10b](https://github.com/yldio/cinzel/commit/5f9c10b676886520cccf3971977defc0ffba4e1f))
- Honor config file and directory inputs - ([731c822](https://github.com/yldio/cinzel/commit/731c822e52ecae406919deabea2f8681bcc7d836))
- Correct outputs mapping - ([013d057](https://github.com/yldio/cinzel/commit/013d0573ff17dbc1877a805ea756f3f623f9f307))
- Stabilize parse and unparse behavior - ([f756910](https://github.com/yldio/cinzel/commit/f7569108b0a9a1016b12815861863f00c1a0901a))

### 🚜 Refactor

- Align github with typed schema contracts - ([4c59e91](https://github.com/yldio/cinzel/commit/4c59e915c0042157637fdc67e4bd80e7cc702fd5))

### 📚 Documentation

- Refresh release architecture references - ([ff1df5b](https://github.com/yldio/cinzel/commit/ff1df5b2816b44d3b453244e0d60a9a8f059157f))
- Capture release cask migration fix - ([0a4a71c](https://github.com/yldio/cinzel/commit/0a4a71c54b7bac535f7019b54d9f3fdec6810f0c))
- Update install guide and release learnings - ([b82681b](https://github.com/yldio/cinzel/commit/b82681b93e8dd0da60a2bb1b071e5589dbc80eb7))
- Document style enforcement learnings - ([e8a57a8](https://github.com/yldio/cinzel/commit/e8a57a844dcb13f266d9287dc5e0870787e71d57))
- Clarify README quick-start wording - ([6f56955](https://github.com/yldio/cinzel/commit/6f56955027b71eca715447b4a363c4707206507c))
- Replace project logo asset - ([830ba6a](https://github.com/yldio/cinzel/commit/830ba6a6bdd48159a1ed233c93d0bc02374c34fe))
- Capture license and docs consistency fix - ([374d611](https://github.com/yldio/cinzel/commit/374d61189ec471dc2162ed264841a7a823d9d787))
- Align strict schema guidance across docs - ([561eb72](https://github.com/yldio/cinzel/commit/561eb722be612eb5bc7a3dcaf583cc3230e50005))
- Codify strict schema policy and findings - ([71e050b](https://github.com/yldio/cinzel/commit/71e050bb12846fe2647d6fa6eb615e7e98130729))
- Update gitlab plan checks and links - ([b1ebb3d](https://github.com/yldio/cinzel/commit/b1ebb3ddd484f4ce965be3b2a92274ec5d96d132))
- Add gitlab notes and style guidance - ([dff7b21](https://github.com/yldio/cinzel/commit/dff7b2193f28d51d3920261305422981d5cb3739))
- Add strict schema parity solution note - ([c961580](https://github.com/yldio/cinzel/commit/c96158081387dcd6b51ca5d73ceb6c6bbaef37bd))
- Resolve strict schema plan open questions - ([5a6a720](https://github.com/yldio/cinzel/commit/5a6a720da080d5b34760fa8d84dd508f12d2bd60))
- Expand project documentation - ([6fd9c90](https://github.com/yldio/cinzel/commit/6fd9c90283143ebd0a8036387a9e9ec5dbab177d))
- Add initial project documentation - ([ff8dda8](https://github.com/yldio/cinzel/commit/ff8dda8c5ff84a2c54c68a79e8b340bec2228fa2))

### 🎨 Styling

- Enforce control-flow spacing conventions - ([39fafae](https://github.com/yldio/cinzel/commit/39fafae8aa790607a80c5c1a16a62fb8fb70f0bd))
- Attach comments directly to code - ([3483f10](https://github.com/yldio/cinzel/commit/3483f10feda43382d199c0757148d803337c50b1))
- Trim blank lines at block starts - ([8569b0b](https://github.com/yldio/cinzel/commit/8569b0b77cdb255f05d27e4e13e71e7684f74dd9))
- Apply repo-wide spacing rules - ([cbc6851](https://github.com/yldio/cinzel/commit/cbc685178de8197e15d63eb7d91645b15bbe4957))

### 🧪 Testing

- Add gitlab golden fixtures and benchmarks - ([9c9a8f0](https://github.com/yldio/cinzel/commit/9c9a8f009e571e691ff8ede5b6de91e8e7d12647))
- Improve defaults test (#21) - ([2ce0f06](https://github.com/yldio/cinzel/commit/2ce0f06b51fb60c214d73f8a473e5cd3af9e3f53))
- Improve permissions test (#20) - ([c1269ff](https://github.com/yldio/cinzel/commit/c1269ff982a308c85210af4506c330970dee48af))
- Improve concurrency test (#19) - ([12a2588](https://github.com/yldio/cinzel/commit/12a2588c3ef06639c70ecb91f44fe0ba781c9227))
- Improve outputs test (#18) - ([13f4851](https://github.com/yldio/cinzel/commit/13f4851750e071e313398220abecaca4068b5d05))
- Improve environment test (#17) - ([b3144d2](https://github.com/yldio/cinzel/commit/b3144d2788047f8393be68dd4defcd64298329c5))
- Improve env test (#16) - ([f8e2d33](https://github.com/yldio/cinzel/commit/f8e2d33c35b83d789dba8f255cae6cf51c1f6bff))
- Improve runs-on test (#15) - ([5ca685a](https://github.com/yldio/cinzel/commit/5ca685af81a2840255a73076a61d7ee8586940e0))
- Improve if test (#14) - ([77cf9c7](https://github.com/yldio/cinzel/commit/77cf9c7594016a650153691ae51d76457ad0e4b0))
- Improve name test (#13) - ([d31d10e](https://github.com/yldio/cinzel/commit/d31d10ec6f6d525b62037c8d5d72e71eae2fa22c))
- Improve timeout-minutes test (#12) - ([3d19435](https://github.com/yldio/cinzel/commit/3d194354ee4a3a04ad2cc5132e9fd34ed6cdffb2))
- Improve continue-on-error test (#11) - ([e416e3c](https://github.com/yldio/cinzel/commit/e416e3c396e9f6c73907193c4288a33ab5e7cef4))
- Improve strategy test (#10) - ([1b5e1b6](https://github.com/yldio/cinzel/commit/1b5e1b67f4eac06e30137416899c560f6cc65c65))
- Improve container test (#9) - ([da01c6f](https://github.com/yldio/cinzel/commit/da01c6fb6491186ca79edeccaf1d7a5a0f842acc))
- Improve uses test (#8) - ([0f46f74](https://github.com/yldio/cinzel/commit/0f46f748e24c37e291825e6208b987f9f97d1ed1))
- Improve with test (#7) - ([bae32d6](https://github.com/yldio/cinzel/commit/bae32d6a79bba3d1a1330f82931325da87756171))
- Improve job secrets (#5) - ([e780a78](https://github.com/yldio/cinzel/commit/e780a78614d7741dbe23ce1174c9bc04ea47040f))
- Improve services test (#4) - ([e0c2310](https://github.com/yldio/cinzel/commit/e0c2310e0a8735eca8f601eb8cce4f4c625f247b))

### ⚙️ Miscellaneous Tasks

- Split release dispatch and published workflows - ([b367294](https://github.com/yldio/cinzel/commit/b3672943275966639c8353b18f34890892e01fd8))
- Simplify workflow test step - ([ee9d149](https://github.com/yldio/cinzel/commit/ee9d14986933562fc17d0718be1efc77604b90fe))
- Remove docker tasks and refresh plan - ([d2b0d49](https://github.com/yldio/cinzel/commit/d2b0d49ea569a51d3add140c5277c271adb23969))
- Add manual release workflow automation - ([e89c043](https://github.com/yldio/cinzel/commit/e89c043ca3ae1f9bdd1e64ab28a9e8c0b1d0d9c9))
- Migrate Homebrew release to casks - ([1b3917c](https://github.com/yldio/cinzel/commit/1b3917c4851ddc14b8320abd4cf381159ec29add))
- Migrate license to Apache 2.0 - ([814db52](https://github.com/yldio/cinzel/commit/814db528cb0c2b8538ad82e071b31b1160ef4901))
- Update deps and add release docs - ([e86dbc5](https://github.com/yldio/cinzel/commit/e86dbc57024dc44f852d45287a6336b311e05d82))
- Add local config and planning docs - ([040ee57](https://github.com/yldio/cinzel/commit/040ee57f4e319dda6369dc9385b5bd6e4b57a218))
- Refactor variables - ([f5ce2a3](https://github.com/yldio/cinzel/commit/f5ce2a3822694c86b0c49ba65bbe67df7f3de7a9))
- Apply repository housekeeping changes - ([0107140](https://github.com/yldio/cinzel/commit/010714079ab6b9695fbe09afe6cd9173f4d4759d))
- Update deps - ([a11717b](https://github.com/yldio/cinzel/commit/a11717b58e192230963da9faa21b21d2d7a3edc2))
- Improve project tooling and structure - ([afa891d](https://github.com/yldio/cinzel/commit/afa891d87ca3cd6d4fc066c0b312841c96addd1f))
- Rebrand (#24) - ([ec4b5e2](https://github.com/yldio/cinzel/commit/ec4b5e2368d1ef0bc348d4ec7176416f0583a30f))
- Refactor and improvements (#22) - ([d71c90c](https://github.com/yldio/cinzel/commit/d71c90c3b0fca9401dffc72d1b4cfbed0ff2f8e6))

### Build

- Migrate changelog tooling to git-cliff - ([fb9be4d](https://github.com/yldio/cinzel/commit/fb9be4d3b5164fd30a9d0869d9f21f135ce7dbc3))


<!-- generated by git-cliff -->
