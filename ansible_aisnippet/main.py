from pathlib import Path
from ansible_aisnippet import __version__
from ansible_aisnippet.aisnippet import aisnippet
from typing import Optional
from ansible_aisnippet.helpers import load_yaml
from ansible_aisnippet.helpers import save_yaml_to_file
from ansible_aisnippet.providers.factory import ProviderFactory
from rich.console import Console

from .helpers import convert_to_yaml

import typer
import os
import sys
import shutil


app = typer.Typer()

def _version_callback(value: bool) -> None:
    if value:
        versionText = typer.style(
            f"ansible-aisnippet v{__version__}",
            fg=typer.colors.RED,
            bg=typer.colors.WHITE,
        )
        typer.echo(versionText)
        raise typer.Exit()

@app.command()
def generate(
    text: Optional[str] = typer.Argument(default="Install package htop", help="A description of task to get"),
    verbose: Optional[bool] = typer.Option(
        False, "--verbose", "-v", help="verbose mode"
    ),
    filetasks: Optional[Path] = typer.Option(None, "--filetasks", "-f", exists=True),
    outputfile: Optional[Path] = typer.Option(None, "--outputfile", "-o", exists=False),
    playbook: Optional[bool] = typer.Option(
        False, "--playbook", "-p", help="Create a playbook"
    ),
    provider: Optional[str] = typer.Option(
        None,
        "--provider",
        help=(
            "AI provider to use. Overrides AI_PROVIDER env var. "
            "Available: openai, anthropic, google, azure, mistral, cohere, "
            "ollama, lmstudio, llama, huggingface."
        ),
    ),
    config_file: Optional[Path] = typer.Option(
        None,
        "--config",
        "-c",
        exists=True,
        help="Path to a YAML configuration file.",
    ),
):
    """
    Ask an AI provider to write an Ansible task using a template.
    """
    from ansible_aisnippet.config import Config

    # Build configuration
    if config_file is not None:
        cfg = Config.from_file(str(config_file))
    else:
        cfg = Config.from_env()

    if provider is not None:
        cfg.provider = provider

    # Generate tasks from a file of tasks (yaml)
    if filetasks is not None:
        assistant = aisnippet(verbose=verbose, config=cfg)
        tasks = load_yaml(filetasks)
        output_tasks = assistant.generate_tasks(tasks)
        if playbook:
            output = [
                    {
                        "name": "Playbook generated with AI",
                        "hosts": "all",
                        "gather_facts": True,
                        "tasks": output_tasks
                    }
            ]
        else:
            output = output_tasks
        if outputfile is not None:
            save_yaml_to_file(outputfile, output)
        else:
            console = Console()
            console.print("Result: \n", style="red")
            console.print(convert_to_yaml(output))
    # Generate a single task from a sentence
    else:
        console = Console()
        assistant = aisnippet(verbose=verbose, config=cfg)
        task = assistant.generate_task(text)
        console.print("Result: \n", style="red")
        console.print(convert_to_yaml(task))


@app.command(name="list-providers")
def list_providers():
    """List all available AI providers."""
    console = Console()
    providers = ProviderFactory.list_providers()
    console.print("Available AI providers:", style="bold green")
    for p in providers:
        console.print(f"  • {p}")


@app.callback()
def main(
    version: Optional[bool] = typer.Option(
        None,
        "--version",
        "-v",
        help="Show the application's version and exit.",
        callback=_version_callback,
        is_eager=True,
    ),
) -> None:
    return


if __name__ == "__main__":
    app()
