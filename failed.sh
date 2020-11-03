kdeletefailed () {
	IFS=$(echo -en "\n\b")
	for po in $(kubectl get po --all-namespaces) #| grep -v Running )
	do
	#	kubectl delete po --wait=false -n $(
		echo $po | awk '{print $1}'
       		echo $po | awk '{print $2}'
	done
}

kdeletefailed