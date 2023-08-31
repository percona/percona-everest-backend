## Limitations and boundaries
Everest product has different dependencies and right now we have the following list

- Kubernetes version
- Operator Lifecycle Manager version
- Versions of upstream and monitoring operators (PXC operator, PSMDB operator, PG operator, and Victoria Metrics Operator)
- PMM version

While working on a feature or improvement you should keep in mind the dependencies of external components

## Kubernetes versioning and limitations

1. We should ensure that Everest supports the same versions of Kubernetes clusters as upstream operators support.
2. Kubernetes has EOL dates. However, the EOL date of Kubernetes for Everest is the latest version taken across three cloud providers (AWS, GCP, Azure).
3. Using Kubernetes features should not limit us to any version of Kubernetes (E.g. Validation rules are available from 1.25, however, we still need to support 1.24 and in that case, we should wait until 1.24 EOL)
4. We should keep our operator in Namespaced mode as long as we can. 


## Best practices we follow

We follow these best practices

- [Go best practices](./go_best_practices.md)
- [Operator best practices](./operator_best_practice.md)

## Team principles
1. **We will not ship garbage.** All the code should be tested before we ship it!
2. **Stable Productivity**
    * We will not allow our project to become unstable in productivity. We know that the system has become unstable when every task becomes more complicated and takes more time.
    * The phrase “This system needs to be redesigned” is an admission of failure, we don’t want a new system that does exactly the same as the old one”.
3. **Inexpensive Adaptability.** Software should be easy to change, that should be the first priority. A system that works but can’t be changed becomes obsolete. A system that doesn’t work but can be easily changed, can be easily fixed.
4. **Continuous Improvement.** Everything should get better with time.
5. **Fearless Competence.** Don’t be afraid to change the code. Testing allows you to make fearless changes.
6. **QA will find nothing!** We will make sure the code is tested before it reaches QA, to speed up this process.
7. **We cover for each other.** We will behave like a team, we have to work together and code with each other. In case someone can’t work another can take over.
8. We follow [80/20 rule](https://en.wikipedia.org/wiki/Pareto_principle) meaning that 80% of users will benefit from 20% of features.
9. We follow KISS and YAGNI
10. We do not optimize something prematurely
