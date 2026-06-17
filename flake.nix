{
  description = "ReelSieve RSS filter service";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    { nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        version = "0.1";
        reelsieve = pkgs.buildGo126Module {
          pname = "reelsieve";
          inherit version;
          src = ./.;
          vendorHash = "sha256-udFoywpttlhoOUwiy65g7Y5jhx+TcnHywbiYsEUk5vI=";
          proxyVendor = true;
          subPackages = [ "cmd/reelsieve" ];
          ldflags = [
            "-s"
            "-w"
            "-X main.version=${version}"
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
