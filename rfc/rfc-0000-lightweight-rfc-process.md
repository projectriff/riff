# RFC-000: Lightweight RFC Process for riff

**To be Reviewed By:** Nov 15th 2019

**Authors:** Swapnil Bawaskar

**Status:** **Active** | Accepted | Dropped | Superseded

**Pull Request URL:** https://github.com/projectriff/riff/pull/1358

**Superseded by:** N/A

**Supersedes:** N/A

**Related:** N/A


## Problem
As the scope of the project expands and the number of moving pieces increases, we a way for contributors to propose designs for significant changes, and encourage discussion and feedback in a PR forum before proceeding with implementation.

This document will refer to that process as the “Request For Comments (RFC) process”.

## Solution
The proposed solution to address the problem described above is to have an individual author or a group of authors submit a proposal to the community in the form of a RFC PR with a RFC markdown file, both called RFC-xxx-<some-name>.md, in order to gather feedback and achieve consensus. The RFC follows the same format as used by this proposal.

Much inspiration for this proposal has been drawn from [Apache Geode's RFC process](https://cwiki.apache.org/confluence/display/GEODE/Lightweight+RFC+Process), which in turn draws from Phil Calçado’s [Structured RFC Process](https://philcalcado.com/2018/11/19/a_structured_rfc_process.html).

All RFCs are submitted via PRs to the [github.com/projectriff/riff](https://github.com/projectriff/riff) repo which will be merged when:
- The `Reviewed By` date has elapsed
- The RFC author has summarized the discussion in the PR thread.
- consensus has been reached about the state ("Accepted/Dropped").

### Collaboration
Comments and feedback should be provided on the PR.

Authors should address all comments by the deadline. This doesn't mean every comment and suggestion must be accepted and incorporated, but they should be carefully read and responded to.

Every RFC is in one of the following phases:
* **Active**: The RFC has been drafted and is open for comments. When making the PR, the doc should start out in this state.
* **Accepted**: All the comments have been addressed and the proposed changes have been agreed upon. The implementation may start after this point.
* **Dropped**: The changes proposed on this RFC were not agreed upon and no implementation will follow.
* **Superseded**: The changes proposed on this RFC aren't in effect anymore, the document may be kept for historical purposes and there is a new RFC that’s more current.

### Approval
The proposal should be posted with a date by which the author would like to have the approval decision. How much time is given to comment depends on the size and complexity of the proposed changes, but ideally should be from 2 days to a week.

### When to write an RFC?
There is always the option of going straight to a pull request. However, for larger changes, it is advisable to reduce the risk of rejection of the pull request by first gathering input from the community.

It’s encouraged to write an RFC for any major change. A major change might be:
* Addition of new feature or subsystem
* Changes that impact existing, or introduce new, public APIs.
* Changes that will introduce user-facing configuration or concepts.
* Changes that need to be coordinated across multiple repositories.

### How to write an RFC?
1. Copy the RFC template (in same folder as this document) and write your proposal! It's up to the author's discretion to decide which sections in the template make sense for their proposal. Cover the problem the proposal is solving, who it affects, how you’re proposing to solve it, and answers to frequently asked questions. Explicitly listing the goals will also make it easier to evaluate whether the proposal was successful.
2. Add your RFC to the riff/rfc source directory, update to the next unique number. 
3. Post a PR for your RFC prefixing the title with `RFC-#`, where `#` is the number of your RFC. Make sure to state the deadline clearly in your PR. 
4. If the proposed RFC replaces another, update the *Supersedes* field.
5. Answer questions and concerns on the PR. Consider adding questions that get asked more than once to the FAQ section of the RFC.
6. After the deadline for feedback has been reached summarize the consensus and your decision on the PR thread. 
    1. Add a link to the Pull request in the RFC so that the discussion and summary does not get lost.
    2. If the decision is to drop the proposal, the status should be updated to *Dropped*. 
    3. If we are moving forward the status goes to *Accepted*
    4. When there is a newer RFC that replaces this one the status goes to *Superseded* and the *Superseded By* gets updated with the number of the new RFC.
7. Someone with merge privilege can then merge the PR.

### Humble Advice
Some things can be helpful to keep in mind when writing technical documents:

1. Keep the document brief but complete. People don’t have time to thoroughly read and think about extremely long documents and they’ll receive less feedback compared with a shorter document. If you find it challenging to meet this limit then maybe the proposal is too big and could be broken up.
2. Include evidence of the problem if at all possible, even if it’s anecdotal. This can help others see the core causes of the issue rather than only being able to comment on the diagnosis or solution. For example, consider linking to evidence, brief inline quotes, and/or footnotes.
3. IETF RFCs you may see contain strict rules conveyed within the semantic meaning of *SHOULD*, *MUST*, and *MAY*. You don’t need to stress about the particulars of language or semantics when writing riff-RFCs. Focus on explaining your problem and proposal clearly, succinctly, and convincingly rather than going into implementation detail.

## FAQ
