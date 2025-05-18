# FIXME - Integration Tests

## Core
- [x] `core/01_list_fails_no_config`

## Create
- [x] `create/01_create_simple`
- [x] `create/02_create_with_config`
- [x] `create/02_create_with_subdir`
- [x] `create/03_create_with_flags`
- [x] `create/04_create_fails_if_exists`
- [x] `create/05_create_multi_backend/01_use_explicit_backend_flag`
- [x] `create/05_create_multi_backend/02_use_default_backend`
- [x] `create/05_create_multi_backend/03_error_ambiguous_no_default`
- [x] `create/05_create_multi_backend/04_use_single_available_backend`
- [x] `create/05_create_multi_backend/05_error_backend_not_found`
- [x] `create/05_create_multi_backend/06_error_unsupported_backend_type`
- [x] `create/06_create_with_body_stdin`
- [x] `create/07_create_with_body_flag`
- [x] `create/08_create_with_body_file`
- [x] `create/09_create_error_multiple_body_sources`

## Get
- [x] `get/01_get_single_structured_default`
- [x] `get/02_get_single_json_frontmatter`
- [x] `get/03_get_single_yaml_frontmatter`
- [x] `get/04_get_single_body`
- [x] `get/06_get_multiple_structured`

## Init
- [x] `init/01_init_and_list_with_config`
- [x] `init/02_init_in_subdir`

## Update
- [x] `update/01_update_title`

## Version
- [x] `version/01_version_displays`