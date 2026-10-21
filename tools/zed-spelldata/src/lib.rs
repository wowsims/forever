use zed_extension_api::{self as zed, LanguageServerId, Result};

struct SpelldataExtension;

impl zed::Extension for SpelldataExtension {
    fn new() -> Self {
        SpelldataExtension
    }

    fn language_server_command(
        &mut self,
        _language_server_id: &LanguageServerId,
        worktree: &zed::Worktree,
    ) -> Result<zed::Command> {
        let go = worktree
            .which("go")
            .ok_or_else(|| "wowsims-spelldata: `go` is not on the PATH".to_string())?;
        Ok(zed::Command {
            command: go,
            args: vec![
                "-C".to_string(),
                worktree.root_path(),
                "run".to_string(),
                "./tools/spelldata/launch".to_string(),
                "-lsp".to_string(),
            ],
            env: worktree.shell_env(),
        })
    }
}

zed::register_extension!(SpelldataExtension);
