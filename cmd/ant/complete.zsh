#compdef ant

# zsh ships a _ant for Apache Ant; source this after compinit, or put it in a directory earlier
# in $fpath, to complete this ant instead.
_ant() {
    # The words of this command alone, as typed, with the open quote the cursor may be in.
    local cmdline="${(j: :)words[1,CURRENT-1]} ${compstate[quote]}${PREFIX}"
    local -a reply
    reply=("${(@f)$(COMP_ZSH="$cmdline" ant)}")
    # No output splits into one empty line, which compadd would insert.
    reply=(${reply:#})
    (( ${#reply} )) || return 1

    local -a insert_files display_files insert_dirs display_dirs insert_words display_words
    local line display rest value kind
    for line in "${reply[@]}"; do
        display="${line%%$'\t'*}"
        rest="${line#*$'\t'}"
        value="${rest%%$'\t'*}"
        kind="${rest#*$'\t'}"
        # Outside quotes, paths go in quoted by hand, with -Q, as compadd would quote a leading ~;
        # inside them compadd quotes for the quote.
        if [[ $kind == (file|dir) && -z $compstate[quote] ]]; then
            if [[ $value == '~'* ]]; then
                value="~${(q)value#\~}"
            else
                value="${(q)value}"
            fi
        fi
        case $kind in
            file) display_files+=("$display"); insert_files+=("$value") ;;
            dir)  display_dirs+=("$display");  insert_dirs+=("$value") ;;
            *)    display_words+=("$display"); insert_words+=("$value") ;;
        esac
    done

    # -U: ant has matched the word itself, so zsh takes the values as they are.
    local -a quoted
    [[ -z $compstate[quote] ]] && quoted=(-Q)
    (( ${#insert_words} )) && compadd -U -d display_words -a insert_words
    (( ${#insert_files} )) && compadd -U $quoted -f -d display_files -a insert_files
    # A directory takes a removable slash instead of a space, to go on into it.
    (( ${#insert_dirs} )) && compadd -U $quoted -f -q -S / -d display_dirs -a insert_dirs
    return 0
}

if [ "$funcstack[1]" = "_ant" ]; then
    _ant "$@"
else
    compdef _ant ant
fi
