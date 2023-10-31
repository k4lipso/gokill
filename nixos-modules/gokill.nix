flake: { config, lib, pkgs, self, ... }: 
let
  cfg = config.services.gokill;
  configFile = pkgs.writeText "config.json" (builtins.toJSON cfg.triggers); 
  gokill-pkg = self.packages.x86_64-linux.gokill;
in
{
  options = with lib; {
    services.gokill = {
      enable = mkOption {
        default = false;
        type = types.bool;
        description = mdDoc ''
          Enables gokill daemon
          '';
      };

      triggers = mkOption {
        description = "list of triggers";
        default = [];
        type = with types; types.listOf ( submodule {
          options = {
            type = mkOption {
              type = types.str;
            };

            name = mkOption {
              type = types.str;
            };

            options = mkOption {
              type = types.attrs;
            };

            actions = mkOption {
              description = "list of actions";
              type = with types; types.listOf ( submodule {
                options = {
                  type = mkOption {
                    type = types.str;
                  };

                  options = mkOption {
                    type = types.attrs;
                  };

                  stage = mkOption {
                    type = types.int;
                  };
                };
              });
            };
          };
        });
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
}


