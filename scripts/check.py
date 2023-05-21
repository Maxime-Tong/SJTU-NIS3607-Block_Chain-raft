import sys
import re

class node_info:
    def __init__(self, id):
        self.id = id
        self.proposals = []
        self.commits = {}
        self.maxSeq = 0
        self.read_from_log(id)

    def read_from_log(self, id):
        filename = '../log/blockchain' + str(id) + '.log'
        propose_pattern = re.compile("generate Block\[(.*?)\] in seq (\d+) at (\d+)\n")
        commit_pattern = re.compile("commit Block\[(.*?)\] in seq (\d+) at (\d+)\n")
        with open(filename, "r") as f:
            lines = f.readlines()
            for line in lines:
                m1 = propose_pattern.search(line)
                m2 = commit_pattern.search(line)
                if m1 != None:
                    self.proposals.append(m1.group(1))
                if m2 != None:
                    if int(m2.group(2)) not in self.commits.keys():
                        self.maxSeq = max(self.maxSeq, int(m2.group(2)))
                        self.commits[int(m2.group(2))] = m2.group(1)
                    else:
                        if self.commits[int(m2.group(2))] != m2.group(1):
                            print(f"node {id} has two or more different commits at position {m2.group(2)}")
                            sys.exit(-1)
            
            print(f"node {id} commits {len(self.commits.keys())} times")    


def check_safety(nodes):
    n = len(nodes)
    for i in range(n):
        for j in range(n):
            if i < j:
                commits1 = nodes[i].commits
                commits2 = nodes[j].commits
                k = 0
                for k in range(min(nodes[i].maxSeq, nodes[j].maxSeq)):
                    block1 = None
                    if k in commits1.keys():
                        block1 = commits1[k]
                    block2 = None
                    if k in commits2.keys():
                        block2 = commits2[k]
                    if block1 != block2:
                        print(f"node {i} has different commit with node {j} at position {k}, node {i} is {block1} while node {j} is {block2}")
                        return False    
    return True


def check_validity(nodes):
    all_proposals = []
    all_commits = []
    for node in nodes:
        all_proposals += node.proposals
        all_commits += node.commits.values()
    for commit in all_commits:
        if commit not in all_proposals:
             print(f"commit {commit} has never been proposed!")
             return False
    return True



if __name__ == "__main__":
    if len(sys.argv) < 2:
        print("Usage: python3 check.py [node number] [crash node...] ")
        print("\t node number: total node number \n")
        print("\t crash node: 0-2 crash nodes\n") 
        sys.exit()
    n = 0
    try:
        n = int(sys.argv[1])
        crash_nodes = []
        if len(sys.argv) >= 3:
            crash_nodes = [int(x) for x in sys.argv[2:]]    
    except e:
        print("The parameter must be an integer")
        sys.exit()
    print("----------begin to check------------")
    nodes = []
    nodes_for_safety = []

    for i in range(n):
        node = node_info(i)
        nodes.append(node) 
    
    for i in range(n):
        if i not in crash_nodes:
            nodes_for_safety.append(nodes[i]) 


    print("---------- check validity ------------")
    if check_validity(nodes):
        print("pass")
    else:
        print("not pass")

    print("---------- check safety --------------")
    if check_safety(nodes_for_safety):
        print("pass")
    else:
        print("not pass")
    print("---------- end -----------------------")