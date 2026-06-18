{
  description = "ReelSieve RSS filter service";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = builtins.replaceStrings [ "\n" ] [ "" ] (builtins.readFile ./internal/version/VERSION);
        reelsieve = pkgs.buildGo126Module {
          pname = "reelsieve";
          inherit version;
          src = ./.;
          vendorHash = "sha256-Xlgq9cwsOmEWwZdnszZT0IM0ODNfI1NE3JgzHse3grM=";
          subPackages = [ "cmd/reelsieve" ];
          ldflags = [
            "-s"
            "-w"
          ];
        };
      in
      {
        packages = {
          default = reelsieve;
          inherit reelsieve;
          container = pkgs.dockerTools.buildLayeredImage {
            name = "reelsieve";
            tag = version;
            contents = [
              reelsieve
              pkgs.cacert
            ];
            config = {
              Entrypoint = [ "${reelsieve}/bin/reelsieve" ];
              Env = [ "SSL_CERT_FILE=${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt" ];
              ExposedPorts = {
                "8080/tcp" = { };
              };
              User = "65532:65532";
            };
          };
        };
      }
    );
}
