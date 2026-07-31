1. **Create `config.go`**
   - Introduce `LoadConfig` function that reads from `.git-tag-inc.json` or `.gittaginc.json`.
   - Store configurations in `ConfiguredEnvsMap` and `ConfiguredStagesMap`.

2. **Update `tag.go` Structs and Parsing**
   - Replace `Test` and `Uat` fields with `EnvName string` and `Env *int` in `Tag` struct.
   - Update `ParseTag` to loop and use the new maps for detecting stages and envs. This replaces the regex parsing logic and supports arbitrary orders/suffixes.
   - Update `Tag.LessThan`, `Tag.String`, `Tag.Clone`, `Tag.CopyFrom`, `Tag.applyIncrement`, and `envInfo` to use `EnvName` and `Env`.
   - Update `detectDecreases` logic for the new fields.

3. **Remove `regexp` from `util.go`**
   - Rewrite `CommandsToFlags` to parse suffixes using a manual loop, checking `ConfiguredStagesMap` and `ConfiguredEnvsMap`.

4. **Update Tests (`tag_test.go`)**
   - Update test expected structs to use `EnvName` and `Env`.
   - Validate tests pass with the new configurable suffix architecture.

5. **Complete pre-commit steps**
   - Ensure proper testing, verification, review, and reflection are done.

6. **Submit Code**
   - Request review and submit.
