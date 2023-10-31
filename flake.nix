{
  description = "A very basic flake";

  #nixpkgs for testing framework
  inputs.nixpkgs.url = "github:nixos/nixpkgs/nixpkgs-unstable";

  outputs = { self, nixpkgs, ... }: 
  let
    forAllSystems = nixpkgs.lib.genAttrs [ "x86_64-linux" ];
    pkgs = nixpkgs.legacyPackages."x86_64-linux";
  in
  {
    devShell."x86_64-linux" = pkgs.mkShell {
      packages = with pkgs; [
        go
        gotools
        mdbook
      ];
    };

    packages.x86_64-linux.gokill = nixpkgs.legacyPackages.x86_64-linux.buildGoModule rec {
      pname = "gokill";
      version = "1.0";
      vendorHash = "sha256-aKEOMeW9QVSLsSuHV4b1khqM0rRrMjJ6Eu5RjY+6V8k=";
      src = ./.;

      postInstall = ''
        '';
    };

    packages.x86_64-linux.gokill-docbuilder = nixpkgs.legacyPackages.x86_64-linux.buildGoModule rec {
      pname = "docbuilder";
      version = "1.0";
      vendorHash = null;
      buildFLags = "-o . $dest/cmd/gokill/docbuilder";
      src = ./.;

      postInstall = ''
        '';
    };


    packages.x86_64-linux.docs = pkgs.callPackage (import ./docs/default.nix) { self = self; };

    packages.x86_64-linux.default = self.packages.x86_64-linux.gokill;

    nixosModules.gokill = { config, lib, pkgs, ... }: 
    let
      cfg = config.services.gokill;
      configFile = pkgs.writeText "config.json" (builtins.toJSON cfg.triggers); 
      gokill-pkg = self.packages.x86_64-linux.gokill;
    in
    {
      options = {
        services.gokill = {
          enable = lib.mkOption {
            default = false;
            type = lib.types.bool;
            description = lib.mdDoc ''
              Enables gokill daemon
              '';
          };

          triggers = lib.mkOption {
            description = "list of triggers";
            default = [];
            type = with lib.types; lib.types.listOf ( submodule {
              options = {
                type = lib.mkOption {
                  type = lib.types.str;
                };

                name = lib.mkOption {
                  type = lib.types.str;
                };

                options = lib.mkOption {
                  type = lib.types.attrs;
                };

                actions = lib.mkOption {
                  description = "list of actions";
                  type = with lib.types; lib.types.listOf ( submodule {
                    options = {
                      type = lib.mkOption {
                        type = lib.types.str;
                      };

                      options = lib.mkOption {
                        type = lib.types.attrs;
                      };

                      stage = lib.mkOption {
                        type = lib.types.int;
                      };
                    };
                  });
                };
              };
            });
          };

          extraConfig = lib.mkOption {
            type = lib.types.str;
            description = lib.mdDoc ''
              gokill config.json
              '';
          };
        };
      };

      config = lib.mkIf cfg.enable {
        systemd.services.gokill = {
          description = "gokill daemon";
          serviceConfig = {
            Type = "simple";
            ExecStart = "${gokill-pkg}/bin/gokill -c ${configFile}";
            Restart = "on-failure";
          };

          wantedBy = [ "default.target" ];
        };
      };
    };

    packages.x86_64-linux.testVm = 
    let
      nixos = nixpkgs.lib.nixosSystem {
        system = "x86_64-linux";
        modules = [
          self.nixosModules.gokill
          {
            services.gokill.enable = true;
            services.gokill.triggers = [
              {
                type = "Timeout";
                name = "custom timeout";
                options = {
                  duration =  10;
                };
                actions = [
                    {
                        type = "Timeout";
                        options = {
                          duration = 5;
                        };
                        stage = 1;
                    }
                    {
                        type = "Shutdown";
                        options = {
                        };
                        stage = 2;
                    }
                ];
              }
            ];
            users.users.root.password = "root";
            virtualisation.vmVariant.virtualisation.graphics = false;
          }
        ];

      };
    in
    nixos.config.system.build.vm;

    apps.x86_64-linux.testVm = {
      type = "app";
      program = builtins.toString (nixpkgs.legacyPackages."x86_64-linux".writeScript "vm" ''
        ${self.packages."x86_64-linux".testVm}/bin/run-nixos-vm
      '');
    };

    apps.x86_64-linux.docs = {
      type = "app";
      program = builtins.toString (nixpkgs.legacyPackages."x86_64-linux".writeScript "docs" ''
        ${pkgs.python3}/bin/python3 -m http.server --directory ${self.packages."x86_64-linux".docs}/share/doc'');
    };

    checks = forAllSystems (system: let
      checkArgs = {
        pkgs = nixpkgs.legacyPackages.${system};
        inherit self;
      };
    in {
      gokill = import ./test/test.nix checkArgs;
    });
  };
}
