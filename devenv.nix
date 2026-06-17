{ pkgs, ... }:
{
  packages = [ pkgs.libxml2 ];

  enterShell = ''
    echo "🛠️ reelsieve dev shell"
  '';

  # See full reference at https://devenv.sh/reference/options/
}
