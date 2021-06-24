{
    min_cov_level=int(min_cov_level)
    curr_cov_level=int($1)
}
{
    if (curr_cov_level < min_cov_level) {
        printf "Coverage level %d%% is below minimum %d%%\n", curr_cov_level, min_cov_level
        exit 1
    }
}
