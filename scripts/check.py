import sys
import re

class node_info:
    def __init__(self, id):
        self.id = id
        self.proposals = {}
        self.commits = []
        self.read_from_log(id)

    def read_from_log(self, id):
        filename = '../log/node' + str(id) + '.log'
        propose_pattern = re.compile("generated Block\[(.*?)\] at (\d+)\n")
        commit_pattern = re.compile("committed Block\[(.*?)\] at (\d+)\n")
        latencies = []
        with open(filename, "r") as f:
            lines = f.readlines()
            for line in lines:
                m1 = propose_pattern.search(line)
                m2 = commit_pattern.search(line)
                if m1 != None:
                    self.proposals[m1.group(1)] = int(m1.group(2))
                if m2 != None:
                    self.commits.append(m2.group(1))
                    if m2.group(1) in self.proposals.keys():
                        # print(m2.group(1), " ", int(m2.group(2)), " ", self.proposal[m2.group(1)])
                        latencies.append(int(m2.group(2)) - self.proposals[m2.group(1)])
            if len(latencies) !=  0:
                print(f"node {id}: propose {len(self.proposals.keys())} times; commit {len(self.commits)} times; average latency is {sum(latencies)/len(latencies)} ns")
            else:
                print(f"node {id}: propose {len(self.proposals.keys())} times; commit {len(self.commits)} times")

def check_safety(nodes):
    n = len(nodes)
    for i in range(n):
        for j in range(n):
            if i < j:
                commits1 = nodes[i].commits
                commits2 = nodes[j].commits
                for k in range(min(len(commits1),len(commits2))):
                    if commits1[k] != commits2[k]:
                        print(f"node {i} has different commit with node {j} at position {k}, node {i} is {commits1[k]} while node {j} is {commits2[k]}")
                        return False    
    return True
def check_validity(nodes):
    all_proposals = []
    all_commits = []
    for node in nodes:
        all_proposals += node.proposals.keys()
        all_commits += node.commits
    for commit in all_commits:
        if commit not in all_proposals:
             print(f"commit {commit} has never been proposed!")
             return False
    return True



if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python3 check.py [node number]")
        sys.exit()
    n = 0
    try:
        n = int(sys.argv[1])
    except e:
        print("The parameter must be an integer")
        sys.exit()
    print("----------begin to check------------")
    nodes = []
    for i in range(n):
        node = node_info(i)
        nodes.append(node) 

    print("---------- check validity ------------")
    if check_validity(nodes):
        print("pass")
    else:
        print("not pass")

    print("---------- check safety --------------")
    if check_safety(nodes):
        print("pass")
    else:
        print("not pass")
    print("---------- end -----------------------")