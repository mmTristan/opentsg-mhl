#  clog> build
# short> build & inject metadata into clog
# extra> we only check the tags and esure local & remote match
#                             _                               _      _
#   ___   _ __   ___   _ _   | |_   ___  __ _   ___   _ __   | |_   | |
#  / _ \ | '_ \ / -_) | ' \  |  _| (_-< / _` | |___| | '  \  | ' \  | |
#  \___/ | .__/ \___| |_||_|  \__| /__/ \__, |       |_|_|_| |_||_| |_|
#        |_|                            |___/

[ -f clogrc/common.sh ] && source clogrc/common.sh  # helper functions
# -----------------------------------------------------------------------------

source clogrc/check.sh  ignore                    # preflight - ignore warnings
printf "${cT}Project$cS $PROJECT$cX\n"

# --- update local & remote tags ----------------------------------------------

[[ ( "$vLOCAL" == "$vREF" ) && ( "$vRmhl" == "$vREF" ) ]] && exit 0

# offer to tag local repo (default = no)

if [[ "$vLOCAL" != "$vREF" ]] ; then
  fPrompt "${cT}Tag$cS $PROJECT$cT locally @ $vREF?$cX" "yN" 6
  if [ $? -eq 0 ] ; then # yes was selected
    printf "Tagging local with $vREF.\n"
    fTagLocal "$vREF" "matching tag to release ($vREF)"
    [ $? -gt 0 ] && printf "${cE}Abort$cX\n" && exit 1
    vLOCAL=$(git tag | tail -1)
  fi
fi

if [[ ( "$vLOCAL" == "$vREF" ) && ( "$vRmhl" != "$vREF" ) ]] ; then
  fPrompt "${cT}Push$cS $PROJECT$cT to origin @ $vREF?$cX" "yN" 6
  if [ $? -eq 0 ] ; then # yes was selected
    printf "Pushing $vREF to origin.\n"
    fTagRemote "$vREF"
    [ $? -gt 0 ] && printf "${cE}Abort$cX\n" && exit 1
  fi
fi
