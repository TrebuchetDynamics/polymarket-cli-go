#!/usr/bin/env python3
from __future__ import annotations

import argparse
import getpass
import os
from pathlib import Path

from cryptography.fernet import Fernet

from bot.core.crypto import material_to_fernet_key


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Encrypt a private key using the POLY_ENC_KEY Fernet secret."
    )
    parser.add_argument(
        "-i",
        "--input",
        type=str,
        help="Path to plaintext private key file (if omitted, prompt securely).",
    )
    parser.add_argument(
        "-o",
        "--output",
        type=str,
        required=True,
        help="Destination path for the encrypted output (e.g. secrets/l1.enc).",
    )
    parser.add_argument(
        "--env-var",
        type=str,
        default="POLY_ENC_KEY",
        help="Environment variable that stores the Fernet key (default: POLY_ENC_KEY).",
    )
    parser.add_argument(
        "--salt-env",
        type=str,
        default="POLY_ENC_SALT",
        help="Environment variable that stores the derivation salt when using a passphrase.",
    )
    parser.add_argument(
        "--force",
        action="store_true",
        help="Overwrite the output file if it already exists.",
    )
    parser.add_argument(
        "--delete-input",
        action="store_true",
        help="Delete the plaintext input file after encryption succeeds.",
    )
    parser.add_argument(
        "--no-confirm",
        action="store_true",
        help="Skip the double-entry prompt when reading from stdin.",
    )
    return parser.parse_args()


def load_master_key(env_var: str, salt_env: str) -> Fernet:
    key = os.getenv(env_var)
    if not key:
        raise SystemExit(f"Missing environment variable: {env_var}")
    salt = os.getenv(salt_env)
    try:
        derived = material_to_fernet_key(key, salt)
        return Fernet(derived)
    except Exception as exc:  # noqa: BLE001
        raise SystemExit(
            f"Failed to derive Fernet key from {env_var}. Provide a valid key or set salt via {salt_env}."
        ) from exc


def read_plaintext(args: argparse.Namespace) -> str:
    if args.input:
        data = Path(args.input).read_text().strip()
        if not data:
            raise SystemExit("Input file is empty")
        return data

    secret = getpass.getpass("Enter L1 private key: ").strip()
    if not secret:
        raise SystemExit("Empty key provided")
    if not args.no_confirm:
        confirm = getpass.getpass("Re-enter key to confirm: ").strip()
        if secret != confirm:
            raise SystemExit("Inputs do not match")
    return secret


def main() -> None:
    args = parse_args()
    fernet = load_master_key(args.env_var, args.salt_env)
    plaintext = read_plaintext(args)
    ciphertext = fernet.encrypt(plaintext.encode()).decode()

    output_path = Path(args.output).expanduser()
    output_path.parent.mkdir(parents=True, exist_ok=True)
    if output_path.exists() and not args.force:
        raise SystemExit(f"{output_path} already exists (use --force to overwrite)")

    output_path.write_text(ciphertext)
    print(f"Encrypted key written to {output_path}")

    if args.delete_input and args.input:
        Path(args.input).unlink()
        print(f"Deleted plaintext input file {args.input}")


if __name__ == "__main__":
    main()
