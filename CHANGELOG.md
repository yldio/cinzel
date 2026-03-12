cinzel
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
