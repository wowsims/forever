use zed_extension_api::{self as zed, serde_json, settings::LspSettings, LanguageServerId, Result};

struct SimlExtension;

impl zed::Extension for SimlExtension {
    fn new() -> Self {
        SimlExtension
    }

    // The launcher regenerates stale protos, builds wowsimcli from the worktree and runs `wowsimcli siml lsp`.
    fn language_server_command(
        &mut self,
        _language_server_id: &LanguageServerId,
        worktree: &zed::Worktree,
    ) -> Result<zed::Command> {
        let go = worktree
            .which("go")
            .ok_or_else(|| "wowsims-siml: `go` is not on the PATH".to_string())?;
        Ok(zed::Command {
            command: go,
            args: vec![
                "-C".to_string(),
                worktree.root_path(),
                "run".to_string(),
                "./sim/siml/lsp/launch".to_string(),
            ],
            env: worktree.shell_env(),
        })
    }

    // The server's sim commands (a code action and a code lens that sim the rotation, and its rows' counts as inlay
    // hints) are on, under what the user's `lsp.wowsims-siml.initialization_options` set: `simIterations`,
    // `simSettingsFile`, or `simCommands: false`.
    fn language_server_initialization_options(
        &mut self,
        language_server_id: &LanguageServerId,
        worktree: &zed::Worktree,
    ) -> Result<Option<serde_json::Value>> {
        let mut options = serde_json::json!({ "simCommands": true });
        let user = LspSettings::for_worktree(language_server_id.as_ref(), worktree)
            .ok()
            .and_then(|settings| settings.initialization_options);
        if let (Some(serde_json::Value::Object(user)), Some(options)) =
            (user, options.as_object_mut())
        {
            options.extend(user);
        }
        Ok(Some(options))
    }
}

zed::register_extension!(SimlExtension);
