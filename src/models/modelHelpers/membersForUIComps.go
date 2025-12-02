package modelHelpers

import (
	"context"
	"i9chat/src/appTypes/UITypes"
	"runtime"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
)

/*
--- Explaining the memberData for UIComp swaping job sharing between threads ---

  - The goal is to map the list of similar members ([]redis.Z) provided
    to their UI component data populating values as needed
    e.g. members containing postIds are mapped to post ui components

  - Now we don't know how many members there will be, and each mapping function
    sends a couple of requests to the redis database, depending on the data needed
    to produce a full UI component from the ID (ID or username) in the member.
    Therefore, a sequential mapping process is not a scalable approach, a more
    optimized approach will be to share the mapping job between threads
    running in parallel, where each thread works on an even portion of the members list.
    Here's how:

  - A corresponding slice of UIComp of a length equal to that of the members ([]redis.Z)
    is created. Then each thread, iterates over its allocated portion of []redis.Z,
    and inserts the generated UIComp into the []UIComp index that corresponds to
    the current index it's working on in []redis.Z.
    So that, a post's ID in ([]redis.Z)[0] has its UI component data in ([]UIComp)[0],
    or a username in ([]redis.Z)[2] has its UI component data in ([]UIComp)[2]

  - By default, the job is shared evenly between numCPUs threads, although the last thread
    can take one extra job, provided the number of jobs is of odd length.
    But, in the case where the number of jobs is less than numCPUs threads,
    the number of threads to use is truncated to the number of jobs to maintain evenness.
    So, in the least case, we hae a number of threads equal to the number of jobs.

  - `start, end := (jobsLen*j)/threadNums, jobsLen*(j+1)/threadNums`
    is how each thread takes an even portion of jobs (except the last, in an odd case)

    j is the position of the next thread starting from 0.
    jobsLen is the total number of jobs being shared by the threads
    threadNums is the number of threads sharing the jobs (numCPUs, by default)

    threadN works from start to end (exclusive),
    threadN+1 works from the threadN's end (it's start)
    to threadN+2's start (it's end) (exclusive),
    taking a number of jobs equal to threadN's

  - Each thread (goroutine) is started by errgroup, which is just a sync.WaitGroup with
    an implementation to terminate all threads (goroutines) when one signals an error,
    which is the exact behaviour we want, seeing this is one unified task
    shared by many independent processors for resource utilization
*/

func ChatIdentMembersForUIChatSnippets(ctx context.Context, chatIdentMembers []redis.Z, clientUsername string) ([]UITypes.ChatSnippet, error) {

	cimsLen := len(chatIdentMembers)

	chatSnippetsAcc := make([]UITypes.ChatSnippet, cimsLen)

	threadNums := min(cimsLen, runtime.NumCPU())

	eg, sharedCtx := errgroup.WithContext(ctx)

	for i := range threadNums {
		eg.Go(func() error {
			j := i
			start, end := (cimsLen*j)/threadNums, cimsLen*(j+1)/threadNums

			for pIndx := start; pIndx < end; pIndx++ {
				chatIdent := chatIdentMembers[pIndx].Member.(string)
				cursor := chatIdentMembers[pIndx].Score

				chatSnippet, err := buildChatSnippetUIFromCache(sharedCtx, clientUsername, chatIdent)
				if err != nil {
					return err
				}

				chatSnippet.Cursor = cursor

				chatSnippetsAcc[pIndx] = chatSnippet
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return chatSnippetsAcc, nil
}

func CHEMembersForUICHEs(ctx context.Context, CHEMembers []redis.Z, chatType string) ([]UITypes.ChatHistoryEntry, error) {
	chemsLen := len(CHEMembers)

	CHEsAcc := make([]UITypes.ChatHistoryEntry, chemsLen)

	threadNums := min(chemsLen, runtime.NumCPU())

	eg, sharedCtx := errgroup.WithContext(ctx)

	for i := range threadNums {
		eg.Go(func() error {
			j := i
			start, end := (chemsLen*j)/threadNums, chemsLen*(j+1)/threadNums

			for pIndx := start; pIndx < end; pIndx++ {
				CHEId := CHEMembers[pIndx].Member.(string)
				cursor := CHEMembers[pIndx].Score

				CHE, err := buildCHEUIFromCache(sharedCtx, CHEId, chatType)
				if err != nil {
					return err
				}

				CHE.Cursor = cursor

				CHEsAcc[pIndx] = CHE
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return CHEsAcc, nil
}
