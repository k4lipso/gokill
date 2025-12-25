{ config, lib, pkgs, inputs, ... }: 
let
  cfg = config.services.gokill;
  configFile = pkgs.writeText "config.json" (builtins.toJSON cfg.triggers); 
  remoteConfigFile = pkgs.writeText "remote-config.json" (builtins.toJSON cfg.remote.config); 
  gokill-pkg = inputs.self.packages.x86_64-linux.gokill;
  testRun = if cfg.testRun then "-t" else "";
  #remoteCfg = if cfg.remote.config != {} then "-remote-config ${remoteConfigFile}" else "";
  remoteCfg = "";
  remote = if cfg.remote.enable then "-r" else "";
  keyAge = if cfg.remote.ageKeyFile != "" then "-key-age ${cfg.remote.ageKeyFile}" else "";
  keyP2p = if cfg.remote.p2pKeyFile != "" then "-key-p2p ${cfg.remote.p2pKeyFile}" else "";
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

      testRun = mkOption {
        default = false;
        type = types.bool;
        description = mdDoc ''
          When set to true gokill is performing a test run
          '';
      };

      remote = {
        enable = mkOption {
          default = false;
          type = types.bool;
          description = mdDoc ''
            Enables gokills remote handler
            '';
        };

        ageKeyFile = mkOption {
          type = with lib.types; either path str;
          default = "";
          description = mdDoc ''
            Path to gokill age key
            '';
        };

        p2pKeyFile = mkOption {
          type = with lib.types; either path str;
          default = "";
          description = mdDoc ''
            Path to gokill p2p key
            '';
        };

        #remoteConfigFile = mkOption {
        #  type = types.str;
        #  default = "";
        #  description = mdDoc ''
        #    Path to gokill remote config
        #    '';
        #};
      };

      #remote.config = {
      #  id = mkOption {
      #    type = types.str;
      #  };
      #  key = mkOption {
      #    type = types.str;
      #  };
      #  groups = mkOption {
      #    description = "gokill remote config";
      #    default = {};
      #    type = with types; types.listOf ( submodule {
      #      options = {
      #        Name = mkOption {
      #          type = types.str;
      #        };
      #        Id = mkOption {
      #          type = types.str;
      #        };
      #        Peers = mkOption {
      #          description = "list of peers";
      #          type = with types; types.listOf ( submodule {
      #            options = {
      #              id = mkOption {
      #                type = types.str;
      #              };
      #              key = mkOption {
      #                type = types.str;
      #              };
      #            };
      #          });
      #        };
      #      };
      #    });
      #  };
      #};

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
    systemd.services.setup-gokill = {
      description = "Initialize gokill directory";
      wantedBy = [ "gokill.service" ];
      unitConfig.ConditionPathExists = "!/etc/gokill/.is_initialized";
      serviceConfig = {
        Type = "oneshot";
      };
      script = ''
        mkdir /etc/gokill
        touch /etc/gokill/.is_initialized
      '';
    };

    systemd.services.gokill = {
      description = "gokill daemon";
      serviceConfig = {
        Type = "simple";
        ExecStart = "${gokill-pkg}/bin/gokill --db /etc/gokill ${remote} ${keyAge} ${keyP2p} ${remoteCfg} -c ${configFile} ${testRun}";
        Restart = "on-failure";
      };

      wantedBy = [ "default.target" ];
    };

    environment.systemPackages = [
      gokill-pkg
    ];
  };
}


