package messages

const (
	AbortContinueGuidance             = "\n\nTo abort, run \"git-town abort\".\nTo continue after having resolved conflicts, run \"git-town continue\".\n"
	AbortNothingToDo                  = "nothing to abort"
	ArgumentUnknown                   = "unknown argument: %q"
	BranchAlreadyExistsLocally        = "there is already a branch %q"
	BranchAlreadyExistsRemotely       = "there is already a branch %q at the \"origin\" remote"
	BranchCheckoutProblem             = "cannot check out branch %q: %w"
	BranchCurrentProblem              = "cannot determine current branch: %w"
	BranchDiffProblem                 = "cannot determine if branch %q has unmerged commits: %w"
	BranchDoesntContainCommit         = "branch %q does not contain commit %q. Found commits %s"
	BranchDoesntExist                 = "there is no branch %q"
	BranchFeatureCannotCreate         = "cannot create feature branch %q: %w"
	BranchLocalSHAProblem             = "cannot determine SHA of local branch %q: %w"
	BranchLocalProblem                = "cannot determine whether the local branch %q exists: %w"
	BrowserOpen                       = "Please open in a browser: %s\n"
	CacheUnitialized                  = "using a cached value before initialization"
	CommitMessageProblem              = "cannot determine last commit message: %w"
	CompletionTypeUnknown             = "unknown completion type: %q"
	ConfigPullbranchStrategyUnknown   = "unknown pull branch strategy: %q"
	ConfigSyncStrategyUnknown         = "unknown sync strategy: %q"
	ConfigRemoveError                 = "unexpected error while removing the 'git-town' section from the Git configuration: %w"
	ContinueSkipGuidance              = "To continue by skipping the current branch, run \"git-town skip\"."
	DiffConflictWithMain              = "conflicts between your uncommmitted changes and the main branch"
	ValueInvalid                      = "invalid value for %s: %q. Please provide either \"yes\" or \"no\""
	ValueGlobalInvalid                = "invalid value for global %s: %q. Please provide either \"true\" or \"false\""
	ConflictDetectionProblem          = "cannot determine conflicts: %w"
	ContinueNothingToDo               = "nothing to continue"
	ContinueUnresolvedConflicts       = "you must resolve the conflicts before continuing"
	DialogOptionNotFound              = "given initial value %q not in given entries"
	DialogCannotReadAuthor            = "cannot read author from CLI: %w"
	DialogCannotReadBranch            = "cannot read branch from CLI: %w"
	DialogCannotReadAnswer            = "cannot read user answer from CLI: %w"
	DialogUnexpectedResponse          = "unexpected response: %s"
	DiffParentNoFeatureBranch         = "you can only diff-parent feature branches"
	DiffProblem                       = "cannot list diff of %q and %q: %w"
	DirCurrentProblem                 = "cannot determine the current directory"
	FileContentInvalidJSON            = "cannot parse JSON content of file %q: %w"
	FileDeleteProblem                 = "cannot delete file %q: %w"
	FileReadProblem                   = "cannot read file %q: %w"
	FileStatProblem                   = "cannot check file %q: %w"
	FileWriteProblem                  = "cannot write file %q: %w"
	GitUserProblem                    = "cannot determine repo author: %w"
	GitVersionMajorNotNumber          = "cannot convert major version %q to int: %w"
	GitVersionMinorNotNumber          = "cannot convert minor version %q to int: %w"
	GitVersionProblem                 = "cannot determine Git version: %w"
	GitVersionUnexpectedOutput        = "'git version' returned unexpected output: %q.\nPlease open an issue and supply the output of running 'git version'"
	GitVersionTooLow                  = "this app requires Git 2.7.0 or higher"
	HostingBitBucketNotImplemented    = "shipping pull requests via the Bitbucket API is currently not supported. If you need this functionality, please vote for it by opening a ticket at https://github.com/git-town/git-town/issues"
	HostingGitlabMergingViaAPI        = "GitLab API: Merging MR !%d ... "
	HostingGitlabUpdateMRViaAPI       = "GitLab API: Updating target branch for MR !%d to %q ... "
	HostingGiteaNotImplemented        = "shipping pull requests via the Gitea API is currently not supported. If you need this functionality, please vote for it by opening a ticket at https://github.com/git-town/git-town/issues"
	HostingGiteaUpdatePRViaAPI        = "Gitea API: Updating base branch for PR #%d to #%s"
	HostingGithubMergingViaAPI        = "GitHub API: merging PR #%d ... "
	HostingGithubUpdatePRViaAPI       = "GitHub API: updating base branch for PR #%d ... "
	HostingServiceUnknown             = "unknown hosting service: %q"
	InputAddOrRemove                  = `invalid argument %q. Please provide either "add" or "remove"`
	InputYesOrNo                      = `invalid argument: %q. Please provide either "yes" or "no".\n`
	KillOnlyFeatureBranches           = "you can only kill feature branches"
	OfflineNotAllowed                 = "this command requires an active internet connection"
	OpenChangesProblem                = "cannot determine open changes: %w"
	ProposalMultipleFound             = "found %d proposals from branch %q to branch %q"
	ProposalNoNumberGiven             = "no pull request number given"
	ProposalNotFoundForBranch         = "cannot determine proposal for branch %q: %w"
	ProposalTargetBranchUpdateProblem = "cannot update the target branch of proposal %d via the API"
	ProposalURLProblem                = "cannot determine proposal URL from %q to %q: %w"
	RebaseProblem                     = "cannot determine rebase in progress: %w"
	RemoteExistsProblem               = "cannot determine if remote %q exists: %w"
	RemotesProblem                    = "cannot determine remotes: %w"
	RenameBranchNotInSync             = "%q is not in sync with its tracking branch, please sync the branches before renaming"
	RenameMainBranch                  = "the main branch cannot be renamed"
	RenamePerennialBranchWarning      = "%q is a perennial branch. Renaming a perennial branch typically requires other updates. If you are sure you want to do this, use '--force'"
	RenameToSameName                  = "cannot rename branch to current name"
	RepoOutside                       = "this is not a Git repository"
	RunAutoAborting                   = "%s\nAuto-aborting... "
	RunCommandProblem                 = "error running command %q: %w"
	RunstateAbortStepProblem          = "cannot run the abort steps: %w"
	RunstateDeleteProblem             = "cannot delete previous run state: %w"
	RunstateLoadProblem               = "cannot load previous run state: %w"
	RunstateSerializeProblem          = "cannot encode run-state: %w"
	RunstatePathProblem               = "cannot determine the runstate file path: %w"
	RunstateSaveProblem               = "cannot save run state: %w"
	RunstateStepUnknown               = "unknown step type: %q, run \"git town status reset\" to reset it"
	SetParentNoFeatureBranch          = "the branch %q is not a feature branch. Only feature branches can have parent branches"
	ShipAbortedMergeError             = "aborted because commit exited with error"
	ShipBranchNothingToDo             = "the branch %q has no shippable changes"
	ShipNoFeatureBranch               = "the branch %q is not a feature branch. Only feature branches can be shipped"
	ShipOpenChanges                   = "you have uncommitted changes. Did you mean to commit them before shipping?"
	ShippableChangesProblem           = "cannot determine whether branch %q has shippable changes: %w"
	SkipBranchHasConflicts            = "cannot skip branch that resulted in conflicts"
	SkipNothingToDo                   = "nothing to skip"
	SquashCannotReadFile              = "cannot read squash message file %q: %w"
	SquashCommitAuthorProblem         = "error getting squash commit author: %w"
	SquashMessageProblem              = "cannot comment out the squash commit message: %w"
	UndoCreateStepProblem             = "cannot create undo step for %q: %w"
	UndoNothingToDo                   = "nothing to undo"
)
